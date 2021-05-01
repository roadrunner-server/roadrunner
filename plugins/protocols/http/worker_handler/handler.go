package handler

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/protocols/http/config"
)

// MB is 1024 bytes
const MB uint64 = 1024 * 1024

// ErrorEvent represents singular http error event.
type ErrorEvent struct {
	// Request contains client request, must not be stored.
	Request *http.Request

	// Error - associated error, if any.
	Error error

	// event timings
	start   time.Time
	elapsed time.Duration
}

// Elapsed returns duration of the invocation.
func (e *ErrorEvent) Elapsed() time.Duration {
	return e.elapsed
}

// ResponseEvent represents singular http response event.
type ResponseEvent struct {
	// Request contains client request, must not be stored.
	Request *Request

	// Response contains service response.
	Response *Response

	// event timings
	start   time.Time
	elapsed time.Duration
}

// Elapsed returns duration of the invocation.
func (e *ResponseEvent) Elapsed() time.Duration {
	return e.elapsed
}

// Handler serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Handler struct {
	maxRequestSize uint64
	uploads        config.Uploads
	trusted        config.Cidrs
	log            logger.Logger
	pool           pool.Pool
	mul            sync.Mutex
	lsn            events.Listener
}

// NewHandler return handle interface implementation
func NewHandler(maxReqSize uint64, uploads config.Uploads, trusted config.Cidrs, pool pool.Pool) (*Handler, error) {
	if pool == nil {
		return nil, errors.E(errors.Str("pool should be initialized"))
	}
	return &Handler{
		maxRequestSize: maxReqSize * MB,
		uploads:        uploads,
		pool:           pool,
		trusted:        trusted,
	}, nil
}

// AddListener attaches handler event controller.
func (h *Handler) AddListener(l events.Listener) {
	h.mul.Lock()
	defer h.mul.Unlock()

	h.lsn = l
}

// mdwr serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = errors.Op("http_plugin_serve_http")
	start := time.Now()

	// validating request size
	if h.maxRequestSize != 0 {
		const op = errors.Op("http_handler_max_size")
		if length := r.Header.Get("content-length"); length != "" {
			// try to parse the value from the `content-length` header
			size, err := strconv.ParseInt(length, 10, 64)
			if err != nil {
				// if got an error while parsing -> assign 500 code to the writer and return
				http.Error(w, errors.E(op, err).Error(), 500)
				h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, errors.Str("error while parsing value from the `content-length` header")), start: start, elapsed: time.Since(start)})
				return
			}

			if size > int64(h.maxRequestSize) {
				h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, errors.Str("request body max size is exceeded")), start: start, elapsed: time.Since(start)})
				http.Error(w, errors.E(op, errors.Str("request body max size is exceeded")).Error(), 500)
				return
			}
		}
	}

	req, err := NewRequest(r, h.uploads)
	if err != nil {
		// if pipe is broken, there is no sense to write the header
		// in this case we just report about error
		if err == errEPIPE {
			h.sendEvent(ErrorEvent{Request: r, Error: err, start: start, elapsed: time.Since(start)})
			return
		}

		http.Error(w, errors.E(op, err).Error(), 500)
		h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, err), start: start, elapsed: time.Since(start)})
		return
	}

	// proxy IP resolution
	h.resolveIP(req)

	req.Open(h.log)
	defer req.Close(h.log)

	p, err := req.Payload()
	if err != nil {
		http.Error(w, errors.E(op, err).Error(), 500)
		h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, err), start: start, elapsed: time.Since(start)})
		return
	}

	rsp, err := h.pool.Exec(p)
	if err != nil {
		http.Error(w, errors.E(op, err).Error(), 500)
		h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, err), start: start, elapsed: time.Since(start)})
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		http.Error(w, errors.E(op, err).Error(), resp.Status)
		h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, err), start: start, elapsed: time.Since(start)})
		return
	}

	h.handleResponse(req, resp, start)
	err = resp.Write(w)
	if err != nil {
		http.Error(w, errors.E(op, err).Error(), 500)
		h.sendEvent(ErrorEvent{Request: r, Error: errors.E(op, err), start: start, elapsed: time.Since(start)})
	}
}

// handleResponse triggers response event.
func (h *Handler) handleResponse(req *Request, resp *Response, start time.Time) {
	h.sendEvent(ResponseEvent{Request: req, Response: resp, start: start, elapsed: time.Since(start)})
}

// sendEvent invokes event handler if any.
func (h *Handler) sendEvent(event interface{}) {
	if h.lsn != nil {
		h.lsn(event)
	}
}

// get real ip passing multiple proxy
func (h *Handler) resolveIP(r *Request) {
	if h.trusted.IsTrusted(r.RemoteAddr) == false { //nolint:gosimple
		return
	}

	if r.Header.Get("X-Forwarded-For") != "" {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		ipCount := len(ips)

		for i := ipCount - 1; i >= 0; i-- {
			addr := strings.TrimSpace(ips[i])
			if net.ParseIP(addr) != nil {
				r.RemoteAddr = addr
				return
			}
		}

		return
	}

	// The logic here is the following:
	// In general case, we only expect X-Real-Ip header. If it exist, we get the IP address from header and set request Remote address
	// But, if there is no X-Real-Ip header, we also trying to check CloudFlare headers
	// True-Client-IP is a general CF header in which copied information from X-Real-Ip in CF.
	// CF-Connecting-IP is an Enterprise feature and we check it last in order.
	// This operations are near O(1) because Headers struct are the map type -> type MIMEHeader map[string][]string
	if r.Header.Get("X-Real-Ip") != "" {
		r.RemoteAddr = fetchIP(r.Header.Get("X-Real-Ip"))
		return
	}

	if r.Header.Get("True-Client-IP") != "" {
		r.RemoteAddr = fetchIP(r.Header.Get("True-Client-IP"))
		return
	}

	if r.Header.Get("CF-Connecting-IP") != "" {
		r.RemoteAddr = fetchIP(r.Header.Get("CF-Connecting-IP"))
	}
}

package http

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
)

const (
	// EventResponse thrown after the request been processed. See ErrorEvent as payload.
	EventResponse = iota + 500

	// EventError thrown on any non job error provided by road runner server.
	EventError
)

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
	cfg               *Config
	log               *logrus.Logger
	rr                *roadrunner.Server
	mul               sync.Mutex
	lsn               func(event int, ctx interface{})
	internalErrorCode uint64
	appErrorCode      uint64
}

// Listen attaches handler event controller.
func (h *Handler) Listen(l func(event int, ctx interface{})) {
	h.mul.Lock()
	defer h.mul.Unlock()

	h.lsn = l
}

// mdwr serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// validating request size
	if h.cfg.MaxRequestSize != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				h.handleError(w, r, err, start)
				return
			} else if size > h.cfg.MaxRequestSize*1024*1024 {
				h.handleError(w, r, errors.New("request body max size is exceeded"), start)
				return
			}
		}
	}

	req, err := NewRequest(r, h.cfg.Uploads)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	// proxy IP resolution
	h.resolveIP(req)

	req.Open(h.log)
	defer req.Close(h.log)

	p, err := req.Payload()
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	rsp, err := h.rr.Exec(p)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	h.handleResponse(req, resp, start)
	err = resp.Write(w)
	if err != nil {
		h.handleError(w, r, err, start)
	}
}

// handleError sends error.
/*
handleError distinct RR errors and App errors
You can set return distinct error codes for the App and for the RR
*/
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, start time.Time) {
	// if pipe is broken, there is no sense to write the header
	// in this case we just report about error
	if err == errEPIPE {
		h.throw(EventError, &ErrorEvent{Request: r, Error: err, start: start, elapsed: time.Since(start)})
		return
	}
	if errors.Is(err, roadrunner.ErrNoAssociatedPool) ||
		errors.Is(err, roadrunner.ErrAllocateWorker) ||
		errors.Is(err, roadrunner.ErrWorkerNotReady) ||
		errors.Is(err, roadrunner.ErrEmptyPayload) ||
		errors.Is(err, roadrunner.ErrPoolStopped) ||
		errors.Is(err, roadrunner.ErrWorkerAllocateTimeout) ||
		errors.Is(err, roadrunner.ErrAllWorkersAreDead) {
		// for the RR errors, write custom error code
		w.WriteHeader(int(h.internalErrorCode))
	} else {
		// ResponseWriter is ok, write the error code
		w.WriteHeader(int(h.appErrorCode))
	}

	_, err2 := w.Write([]byte(err.Error()))
	// error during the writing to the ResponseWriter
	if err2 != nil {
		// concat original error with ResponseWriter error
		h.throw(EventError, &ErrorEvent{Request: r, Error: errors.New(fmt.Sprintf("error: %v, during handle this error, ResponseWriter error occurred: %v", err, err2)), start: start, elapsed: time.Since(start)})
		return
	}
	h.throw(EventError, &ErrorEvent{Request: r, Error: err, start: start, elapsed: time.Since(start)})
}

// handleResponse triggers response event.
func (h *Handler) handleResponse(req *Request, resp *Response, start time.Time) {
	h.throw(EventResponse, &ResponseEvent{Request: req, Response: resp, start: start, elapsed: time.Since(start)})
}

// throw invokes event handler if any.
func (h *Handler) throw(event int, ctx interface{}) {
	h.mul.Lock()
	defer h.mul.Unlock()

	if h.lsn != nil {
		h.lsn(event, ctx)
	}
}

// get real ip passing multiple proxy
func (h *Handler) resolveIP(r *Request) {
	if !h.cfg.IsTrusted(r.RemoteAddr) {
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
	// In general case, we only expect X-Real-Ip header. If it exist, we get the IP addres from header and set request Remote address
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

package http

import (
	"github.com/pkg/errors"
	"github.com/spiral/roadrunner"
	"net/http"
	"strconv"
	"sync"
)

const (
	// EventResponse thrown after the request been processed. See Event as payload.
	EventResponse = iota + 500

	// EventError thrown on any non job error provided by road runner server.
	EventError
)

// Event represents singular http response event.
type Event struct {
	// Method of the request.
	Method string

	// Uri requested by the client.
	Uri string

	// Status is response status.
	Status int

	// Associated error, if any.
	Error error
}

// Handler serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Handler struct {
	cfg *Config
	rr  *roadrunner.Server
	mul sync.Mutex
	lsn func(event int, ctx interface{})
}

// AddListener attaches pool event watcher.
func (h *Handler) Listen(l func(event int, ctx interface{})) {
	h.mul.Lock()
	defer h.mul.Unlock()

	h.lsn = l
}

// middleware serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// validating request size
	if h.cfg.MaxRequest != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				h.handleError(w, r, err)
				return
			} else if size > h.cfg.MaxRequest*1024*1024 {
				h.handleError(w, r, errors.New("request body max size is exceeded"))
				return
			}
		}
	}

	req, err := NewRequest(r, h.cfg.Uploads)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	if err = req.Open(); err != nil {
		h.handleError(w, r, err)
		return
	}
	defer req.Close()

	p, err := req.Payload()
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	rsp, err := h.rr.Exec(p)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.handleResponse(req, resp)
	resp.Write(w)
}

// handleError sends error.
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	h.throw(EventError, &Event{Method: r.Method, Uri: uri(r), Status: 500, Error: err})

	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

// handleResponse triggers response event.
func (h *Handler) handleResponse(req *Request, resp *Response) {
	h.throw(EventResponse, &Event{Method: req.Method, Uri: req.Uri, Status: resp.Status})
}

// throw invokes event srv if any.
func (h *Handler) throw(event int, ctx interface{}) {
	h.mul.Lock()
	defer h.mul.Unlock()

	if h.lsn != nil {
		h.lsn(event, ctx)
	}
}

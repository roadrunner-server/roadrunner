package http

import (
	"net/http"
	"strconv"
	"github.com/spiral/roadrunner"
	"github.com/pkg/errors"
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

// Server serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Server struct {
	cfg      *Config
	listener func(event int, ctx interface{})
	rr       *roadrunner.Server
}

// AddListener attaches pool event watcher.
func (s *Server) Listen(l func(event int, ctx interface{})) {
	s.listener = l
}

// Handle serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// validating request size
	if s.cfg.MaxRequest != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				s.handleError(w, r, err)
				return
			} else if size > s.cfg.MaxRequest*1024*1024 {
				s.handleError(w, r, errors.New("request body max size is exceeded"))
				return
			}
		}
	}

	req, err := NewRequest(r, s.cfg.Uploads)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	if err = req.Open(); err != nil {
		s.handleError(w, r, err)
		return
	}
	defer req.Close()

	p, err := req.Payload()
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	rsp, err := s.rr.Exec(p)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		s.handleError(w, r, err)
		return
	}

	s.handleResponse(req, resp)
	resp.Write(w)
}

// handleResponse triggers response event.
func (s *Server) handleResponse(req *Request, resp *Response) {
	s.throw(EventResponse, &Event{Method: req.Method, Uri: req.Uri, Status: resp.Status})
}

// handleError sends error.
func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error) {
	s.throw(EventError, &Event{Method: r.Method, Uri: uri(r), Status: 500, Error: err})

	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

// throw invokes event srv if any.
func (s *Server) throw(event int, ctx interface{}) {
	if s.listener != nil {
		s.listener(event, ctx)
	}
}

package http

import (
	"net/http"
	"strconv"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
	"github.com/pkg/errors"
)

// Handler serves http connections to underlying PHP application using PSR-7 protocol. Context will include request headers,
// parsed files and query, payload will include parsed form dataTree (if any).
type Handler struct {
	cfg *Config
	rr  *roadrunner.Server
}

// Handle serve using PSR-7 requests passed to underlying application. Attempts to serve static files first if enabled.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// validating request size
	if h.cfg.MaxRequest != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				h.sendError(w, r, err)
				return
			} else if size > h.cfg.MaxRequest {
				h.sendError(w, r, errors.New("request body max size is exceeded"))
				return
			}
		}
	}

	req, err := NewRequest(r)
	if err != nil {
		h.sendError(w, r, err)
		return
	}

	if err = req.Open(h.cfg); err != nil {
		h.sendError(w, r, err)
		return
	}
	defer req.Close()

	p, err := req.Payload()
	if err != nil {
		h.sendError(w, r, err)
		return
	}

	rsp, err := h.rr.Exec(p)
	if err != nil {
		h.sendError(w, r, err)
		return
	}

	resp, err := NewResponse(rsp)
	if err != nil {
		h.sendError(w, r, err)
		return
	}

	resp.Write(w)
}

// sendError sends error
func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, err error) {
	logrus.Errorf("http: %s", err)
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

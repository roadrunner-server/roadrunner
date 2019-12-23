package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/spiral/roadrunner"
)

// Response handles PSR7 response logic.
type Response struct {
	// Status contains response status.
	Status int `json:"status"`

	// Header contains list of response headers.
	Headers map[string][]string `json:"headers"`

	// associated body payload.
	body interface{}
}

// NewResponse creates new response based on given rr payload.
func NewResponse(p *roadrunner.Payload) (*Response, error) {
	r := &Response{body: p.Body}
	if err := json.Unmarshal(p.Context, r); err != nil {
		return nil, err
	}

	return r, nil
}

// Write writes response headers, status and body into ResponseWriter.
func (r *Response) Write(w http.ResponseWriter) error {
	p, h := handlePushHeaders(r.Headers)
	if pusher, ok := w.(http.Pusher); ok {
		for _, v := range p {
			err := pusher.Push(v, nil)
			if err != nil {
				return err
			}
		}
	}

	h = handleTrailers(h)
	for n, h := range r.Headers {
		for _, v := range h {
			w.Header().Add(n, v)
		}
	}

	w.WriteHeader(r.Status)

	if data, ok := r.body.([]byte); ok {
		_, err := w.Write(data)
		if err != nil {
			return err
		}
	}

	if rc, ok := r.body.(io.Reader); ok {
		if _, err := io.Copy(w, rc); err != nil {
			return err
		}
	}

	return nil
}

func handlePushHeaders(h map[string][]string) ([]string, map[string][]string) {
	var p []string
	pushHeader, ok := h["http2-push"]
	if !ok {
		return p, h
	}

	for _, v := range pushHeader {
		p = append(p, v)
	}

	delete(h, "http2-push")

	return p, h
}

func handleTrailers(h map[string][]string) map[string][]string {
	trailers, ok := h["trailer"]
	if !ok {
		return h
	}

	for _, tr := range trailers {
		for _, n := range strings.Split(tr, ",") {
			n = strings.Trim(n, "\t ")
			if v, ok := h[n]; ok {
				h["Trailer:"+n] = v

				delete(h, n)
			}
		}
	}

	delete(h, "trailer")

	return h
}

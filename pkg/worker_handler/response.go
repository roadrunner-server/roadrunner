package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/payload"
)

// Response handles PSR7 response logic.
type Response struct {
	// Status contains response status.
	Status int `json:"status"`

	// Header contains list of response headers.
	Headers map[string][]string `json:"headers"`

	// associated Body payload.
	Body interface{}
}

// NewResponse creates new response based on given pool payload.
func NewResponse(p payload.Payload) (*Response, error) {
	const op = errors.Op("http_response")
	r := &Response{Body: p.Body}
	if err := json.Unmarshal(p.Context, r); err != nil {
		return nil, errors.E(op, errors.Decode, err)
	}

	return r, nil
}

// Write writes response headers, status and body into ResponseWriter.
func (r *Response) Write(w http.ResponseWriter) error {
	// INFO map is the reference type in golang
	p := handlePushHeaders(r.Headers)
	if pusher, ok := w.(http.Pusher); ok {
		for _, v := range p {
			err := pusher.Push(v, nil)
			if err != nil {
				return err
			}
		}
	}

	handleTrailers(r.Headers)
	for n, h := range r.Headers {
		for _, v := range h {
			w.Header().Add(n, v)
		}
	}

	w.WriteHeader(r.Status)

	if data, ok := r.Body.([]byte); ok {
		_, err := w.Write(data)
		if err != nil {
			return handleWriteError(err)
		}
	}

	if rc, ok := r.Body.(io.Reader); ok {
		if _, err := io.Copy(w, rc); err != nil {
			return err
		}
	}

	return nil
}

func handlePushHeaders(h map[string][]string) []string {
	var p []string
	pushHeader, ok := h[http2pushHeaderKey]
	if !ok {
		return p
	}

	p = append(p, pushHeader...)

	delete(h, http2pushHeaderKey)

	return p
}

func handleTrailers(h map[string][]string) {
	trailers, ok := h[TrailerHeaderKey]
	if !ok {
		return
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

	delete(h, TrailerHeaderKey)
}

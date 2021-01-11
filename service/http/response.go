package http

import (
	"io"
	"net/http"
	"strings"

	j "github.com/json-iterator/go"

	"github.com/spiral/roadrunner"
)

var json = j.ConfigCompatibleWithStandardLibrary

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

	if data, ok := r.body.([]byte); ok {
		_, err := w.Write(data)
		if err != nil {
			return handleWriteError(err)
		}
	}

	if rc, ok := r.body.(io.Reader); ok {
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
	trailers, ok := h[trailerHeaderKey]
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

	delete(h, trailerHeaderKey)
}

package http

import (
	"encoding/json"
	"github.com/spiral/roadrunner"
	"net/http"
	"io"
)

// Response handles PSR7 response logic.
type Response struct {
	// Status contains response status.
	Status int `json:"status"`

	// Headers contains list of response headers.
	Headers map[string][]string `json:"headers"`

	// associated body payload.
	body interface{}
}

// NewResponse creates new response based on given roadrunner payload.
func NewResponse(p *roadrunner.Payload) (*Response, error) {
	r := &Response{body: p.Body}
	if err := json.Unmarshal(p.Context, r); err != nil {
		return nil, err
	}

	return r, nil
}

// Write writes response headers, status and body into ResponseWriter.
func (r *Response) Write(w http.ResponseWriter) error {
	for k, v := range r.Headers {
		for _, h := range v {
			w.Header().Add(k, h)
		}
	}

	w.WriteHeader(r.Status)

	if data, ok := r.body.([]byte); ok {
		w.Write(data)
	}

	if rc, ok := r.body.(io.Reader); ok {
		if _, err := io.Copy(w, rc); err != nil {
			return err
		}
	}

	return nil
}

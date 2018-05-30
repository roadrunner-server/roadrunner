package psr7

import (
	"net/http"
)

type Response struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
}

func (r *Response) Write(w http.ResponseWriter) {
	for k, v := range r.Headers {
		for _, h := range v {
			w.Header().Add(k, h)

		}
	}

	w.WriteHeader(r.Status)
}

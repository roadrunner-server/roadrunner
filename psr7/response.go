package psr7

import (
	"net/http"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
}

func (r *Response) Write(w http.ResponseWriter) {
	push := make([]string, 0)
	for k, v := range r.Headers {
		for _, h := range v {
			if k == "http2-push" {
				push = append(push, h)
			} else {
				w.Header().Add(k, h)
			}
		}
	}

	if p, ok := w.(http.Pusher); ok {
		logrus.Info("PUSH SUPPORTED")
		for _, f := range push {
			logrus.Info("pushing HTTP2 file ", f)
			p.Push(f, nil)
		}
	}

	w.WriteHeader(r.Status)
}

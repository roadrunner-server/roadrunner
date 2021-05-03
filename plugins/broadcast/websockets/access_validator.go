package websockets

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
)

type accessValidator struct {
	buffer *bytes.Buffer
	header http.Header
	status int
}

func newValidator() *accessValidator {
	return &accessValidator{
		buffer: bytes.NewBuffer(nil),
		header: make(http.Header),
	}
}

// copy all content to parent response writer.
func (w *accessValidator) copy(rw http.ResponseWriter) {
	rw.WriteHeader(w.status)

	for k, v := range w.header {
		for _, vv := range v {
			rw.Header().Add(k, vv)
		}
	}

	_, _ = io.Copy(rw, w.buffer)
}

// Header returns the header map that will be sent by WriteHeader.
func (w *accessValidator) Header() http.Header {
	return w.header
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *accessValidator) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *accessValidator) WriteHeader(statusCode int) {
	w.status = statusCode
}

// IsOK returns true if response contained 200 status code.
func (w *accessValidator) IsOK() bool {
	return w.status == 200
}

// Body returns response body to rely to user.
func (w *accessValidator) Body() []byte {
	return w.buffer.Bytes()
}

// Error contains server response.
func (w *accessValidator) Error() string {
	return w.buffer.String()
}

// assertServerAccess checks if user can join server and returns error and body if user can not. Must return nil in
// case of error
func (w *accessValidator) assertServerAccess(f http.HandlerFunc, r *http.Request) error {
	if err := attributes.Set(r, "ws:joinServer", true); err != nil {
		return err
	}

	defer delete(attributes.All(r), "ws:joinServer")

	f(w, r)

	if !w.IsOK() {
		return w
	}

	return nil
}

// assertAccess checks if user can access given upstream, the application will receive all user headers and cookies.
// the decision to authorize user will be based on response code (200).
func (w *accessValidator) assertTopicsAccess(f http.HandlerFunc, r *http.Request, channels ...string) error {
	if err := attributes.Set(r, "ws:joinTopics", strings.Join(channels, ",")); err != nil {
		return err
	}

	defer delete(attributes.All(r), "ws:joinTopics")

	f(w, r)

	if !w.IsOK() {
		return w
	}

	return nil
}

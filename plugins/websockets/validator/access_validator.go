package validator

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/spiral/roadrunner/v2/plugins/channel"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
)

type AccessValidator struct {
	buffer *bytes.Buffer
	header http.Header
	status int
}

func NewValidator() *AccessValidator {
	return &AccessValidator{
		buffer: bytes.NewBuffer(nil),
		header: make(http.Header),
	}
}

// Copy all content to parent response writer.
func (w *AccessValidator) Copy(rw http.ResponseWriter) {
	rw.WriteHeader(w.status)

	for k, v := range w.header {
		for _, vv := range v {
			rw.Header().Add(k, vv)
		}
	}

	_, _ = io.Copy(rw, w.buffer)
}

// Header returns the header map that will be sent by WriteHeader.
func (w *AccessValidator) Header() http.Header {
	return w.header
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *AccessValidator) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *AccessValidator) WriteHeader(statusCode int) {
	w.status = statusCode
}

// IsOK returns true if response contained 200 status code.
func (w *AccessValidator) IsOK() bool {
	return w.status == 200
}

// Body returns response body to rely to user.
func (w *AccessValidator) Body() []byte {
	return w.buffer.Bytes()
}

// Error contains server response.
func (w *AccessValidator) Error() string {
	return w.buffer.String()
}

// AssertServerAccess checks if user can join server and returns error and body if user can not. Must return nil in
// case of error
func (w *AccessValidator) AssertServerAccess(hub channel.Hub, r *http.Request) error {
	if err := attributes.Set(r, "ws:joinServer", true); err != nil {
		return err
	}

	defer delete(attributes.All(r), "ws:joinServer")

	hub.ReceiveCh() <- struct {
		RW  http.ResponseWriter
		Req *http.Request
	}{
		w,
		r,
	}

	resp := <-hub.SendCh()

	rmsg := resp.(struct {
		RW  http.ResponseWriter
		Req *http.Request
	})

	if !rmsg.RW.(*AccessValidator).IsOK() {
		return w
	}

	return nil
}

// AssertTopicsAccess checks if user can access given upstream, the application will receive all user headers and cookies.
// the decision to authorize user will be based on response code (200).
func (w *AccessValidator) AssertTopicsAccess(hub channel.Hub, r *http.Request, channels ...string) error {
	if err := attributes.Set(r, "ws:joinTopics", strings.Join(channels, ",")); err != nil {
		return err
	}

	defer delete(attributes.All(r), "ws:joinTopics")

	hub.ReceiveCh() <- struct {
		RW  http.ResponseWriter
		Req *http.Request
	}{
		w,
		r,
	}

	resp := <-hub.SendCh()

	rmsg := resp.(struct {
		RW  http.ResponseWriter
		Req *http.Request
	})

	if !rmsg.RW.(*AccessValidator).IsOK() {
		return w
	}

	return nil
}

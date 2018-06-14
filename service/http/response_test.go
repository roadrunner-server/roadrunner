package http

import (
	"bytes"
	"errors"
	"github.com/spiral/roadrunner"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type testWriter struct {
	h           http.Header
	buf         bytes.Buffer
	wroteHeader bool
	code        int
	err         error
}

func (tw *testWriter) Header() http.Header { return tw.h }

func (tw *testWriter) Write(p []byte) (int, error) {
	if !tw.wroteHeader {
		tw.WriteHeader(http.StatusOK)
	}

	n, e := tw.buf.Write(p)
	if e == nil {
		e = tw.err
	}

	return n, e
}

func (tw *testWriter) WriteHeader(code int) { tw.wroteHeader = true; tw.code = code }

func TestNewResponse_Error(t *testing.T) {
	r, err := NewResponse(&roadrunner.Payload{Context: []byte(`invalid payload`)})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestNewResponse_Write(t *testing.T) {
	r, err := NewResponse(&roadrunner.Payload{
		Context: []byte(`{"headers":{"key":["value"]},"status": 301}`),
		Body:    []byte(`sample body`),
	})

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Equal(t, 301, w.code)
	assert.Equal(t, "value", w.h.Get("key"))
	assert.Equal(t, "sample body", w.buf.String())
}

func TestNewResponse_Stream(t *testing.T) {
	r, err := NewResponse(&roadrunner.Payload{
		Context: []byte(`{"headers":{"key":["value"]},"status": 301}`),
	})

	r.body = &bytes.Buffer{}
	r.body.(*bytes.Buffer).WriteString("hello world")

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Equal(t, 301, w.code)
	assert.Equal(t, "value", w.h.Get("key"))
	assert.Equal(t, "hello world", w.buf.String())
}

func TestNewResponse_StreamError(t *testing.T) {
	r, err := NewResponse(&roadrunner.Payload{
		Context: []byte(`{"headers":{"key":["value"]},"status": 301}`),
	})

	r.body = &bytes.Buffer{}
	r.body.(*bytes.Buffer).WriteString("hello world")

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string)), err: errors.New("error")}
	assert.Error(t, r.Write(w))
}

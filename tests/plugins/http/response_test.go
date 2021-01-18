package http

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/spiral/roadrunner/v2/pkg/payload"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/stretchr/testify/assert"
)

type testWriter struct {
	h           http.Header
	buf         bytes.Buffer
	wroteHeader bool
	code        int
	err         error
	pushErr     error
	pushes      []string
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

func (tw *testWriter) Push(target string, opts *http.PushOptions) error {
	tw.pushes = append(tw.pushes, target)

	return tw.pushErr
}

func TestNewResponse_Error(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{Context: []byte(`invalid payload`)})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestNewResponse_Write(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{
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
	r, err := httpPlugin.NewResponse(payload.Payload{
		Context: []byte(`{"headers":{"key":["value"]},"status": 301}`),
	})

	// r is pointer, so, it might be nil
	if r == nil {
		t.Fatal("response is nil")
		return
	}

	r.Body = new(bytes.Buffer)
	r.Body.(*bytes.Buffer).WriteString("hello world")

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Equal(t, 301, w.code)
	assert.Equal(t, "value", w.h.Get("key"))
	assert.Equal(t, "hello world", w.buf.String())
}

func TestNewResponse_StreamError(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{
		Context: []byte(`{"headers":{"key":["value"]},"status": 301}`),
	})

	// r is pointer, so, it might be nil
	if r == nil {
		t.Fatal("response is nil")
		return
	}

	r.Body = &bytes.Buffer{}
	r.Body.(*bytes.Buffer).WriteString("hello world")

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string)), err: errors.New("error")}
	assert.Error(t, r.Write(w))
}

func TestWrite_HandlesPush(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{
		Context: []byte(`{"headers":{"Http2-Push":["/test.js"],"content-type":["text/html"]},"status": 200}`),
	})

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Nil(t, w.h["Http2-Push"])
	assert.Equal(t, []string{"/test.js"}, w.pushes)
}

func TestWrite_HandlesTrailers(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{
		Context: []byte(`{"headers":{"Trailer":["foo, bar", "baz"],"foo":["test"],"bar":["demo"]},"status": 200}`),
	})

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Nil(t, w.h[httpPlugin.TrailerHeaderKey])
	assert.Nil(t, w.h["foo"]) //nolint:staticcheck
	assert.Nil(t, w.h["baz"]) //nolint:staticcheck

	assert.Equal(t, "test", w.h.Get("Trailer:foo"))
	assert.Equal(t, "demo", w.h.Get("Trailer:bar"))
}

func TestWrite_HandlesHandlesWhitespacesInTrailer(t *testing.T) {
	r, err := httpPlugin.NewResponse(payload.Payload{
		Context: []byte(
			`{"headers":{"Trailer":["foo\t,bar  ,    baz"],"foo":["a"],"bar":["b"],"baz":["c"]},"status": 200}`),
	})

	assert.NoError(t, err)
	assert.NotNil(t, r)

	w := &testWriter{h: http.Header(make(map[string][]string))}
	assert.NoError(t, r.Write(w))

	assert.Equal(t, "a", w.h.Get("Trailer:foo"))
	assert.Equal(t, "b", w.h.Get("Trailer:bar"))
	assert.Equal(t, "c", w.h.Get("Trailer:baz"))
}

package websockets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWrapper_Body(t *testing.T) {
	w := newValidator()
	_, _ =w.Write([]byte("hello"))

	assert.Equal(t, []byte("hello"), w.Body())
}

func TestResponseWrapper_Header(t *testing.T) {
	w := newValidator()
	w.Header().Set("k", "value")

	assert.Equal(t, "value", w.Header().Get("k"))
}

func TestResponseWrapper_StatusCode(t *testing.T) {
	w := newValidator()
	w.WriteHeader(200)

	assert.True(t, w.IsOK())
}

func TestResponseWrapper_StatusCodeBad(t *testing.T) {
	w := newValidator()
	w.WriteHeader(400)

	assert.False(t, w.IsOK())
}

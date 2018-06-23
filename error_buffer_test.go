package roadrunner

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrBuffer_Write_Len(t *testing.T) {
	buf := &errBuffer{buffer: new(bytes.Buffer)}
	buf.Write([]byte("hello"))
	assert.Equal(t, 5, buf.Len())
	assert.Equal(t, buf.String(), "hello")
}

func TestErrBuffer_Write_Event(t *testing.T) {
	buf := &errBuffer{buffer: new(bytes.Buffer)}

	tr := make(chan interface{})
	buf.Listen(func(event int, ctx interface{}) {
		assert.Equal(t, EventStderrOutput, event)
		assert.Equal(t, []byte("hello"), ctx)
		close(tr)
	})

	buf.Write([]byte("hello"))

	<-tr
	assert.Equal(t, 5, buf.Len())
	assert.Equal(t, buf.String(), "hello")
}

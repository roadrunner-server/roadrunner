package roadrunner

import (
	"testing"
	"bytes"
	"github.com/magiconair/properties/assert"
)

func TestErrBuffer_Write_Len(t *testing.T) {
	buf := &errBuffer{buffer: new(bytes.Buffer)}
	buf.Write([]byte("hello"))
	assert.Equal(t, 5, buf.Len())
	assert.Equal(t, buf.String(), "hello")
}

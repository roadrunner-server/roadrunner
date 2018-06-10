package utils

import (
	"testing"
	"github.com/magiconair/properties/assert"
)

func TestUtils(t *testing.T) {
	assert.Equal(t, int64(1024), ParseSize("1K"))
	assert.Equal(t, int64(1024*1024), ParseSize("1M"))
	assert.Equal(t, int64(2*1024*1024*1024), ParseSize("2G"))
}

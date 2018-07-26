package env

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewService(t *testing.T) {
	s := NewService("version")
	assert.Len(t, s.values, 1)
}

func Test_Extend(t *testing.T) {
	s := NewService("version")

	s.Init(&Config{Values: map[string]string{"key": "value"}})
	assert.Len(t, s.values, 2)

	values, err := s.GetEnv()
	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "version", values["rr"])
	assert.Equal(t, "value", values["key"])
}

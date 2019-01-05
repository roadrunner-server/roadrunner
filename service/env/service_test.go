package env

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewService(t *testing.T) {
	s := NewService(map[string]string{"version": "test"})
	assert.Len(t, s.values, 1)
}

func Test_Init(t *testing.T) {
	s := &Service{}
	s.Init(&Config{})
	assert.Len(t, s.values, 1)

	values, err := s.GetEnv()
	assert.NoError(t, err)
	assert.Equal(t, "true", values["RR"])
}

func Test_Extend(t *testing.T) {
	s := NewService(map[string]string{"RR": "version"})

	s.Init(&Config{Values: map[string]string{"key": "value"}})
	assert.Len(t, s.values, 2)

	values, err := s.GetEnv()
	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "version", values["RR"])
	assert.Equal(t, "value", values["key"])
}

func Test_Set(t *testing.T) {
	s := NewService(map[string]string{"RR": "version"})

	s.Init(&Config{Values: map[string]string{"key": "value"}})
	assert.Len(t, s.values, 2)

	s.SetEnv("key", "value-new")
	s.SetEnv("other", "new")

	values, err := s.GetEnv()
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "version", values["RR"])
	assert.Equal(t, "value-new", values["key"])
	assert.Equal(t, "new", values["other"])
}

func Test_Copy(t *testing.T) {
	s1 := NewService(map[string]string{"RR": "version"})
	s2 := NewService(map[string]string{})

	s1.SetEnv("key", "value-new")
	s1.SetEnv("other", "new")

	assert.NoError(t, s1.Copy(s2))

	values, err := s2.GetEnv()
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "version", values["RR"])
	assert.Equal(t, "value-new", values["key"])
	assert.Equal(t, "new", values["other"])
}

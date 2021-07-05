package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJob_Body(t *testing.T) {
	j := &Job{Payload: "hello"}

	assert.Equal(t, []byte("hello"), j.Body())
}

func TestJob_Context(t *testing.T) {
	j := &Job{Job: "job", Options: &Options{ID: "id"}}

	assert.Equal(t, []byte(`{"id":"id","job":"job"}`), j.Context())
}

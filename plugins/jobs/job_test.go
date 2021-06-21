package jobs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJob_Body(t *testing.T) {
	j := &Job{Payload: "hello"}

	assert.Equal(t, []byte("hello"), j.Body())
}

func TestJob_Context(t *testing.T) {
	j := &Job{Job: "job"}

	assert.Equal(t, []byte(`{"id":"id","job":"job"}`), j.Context("id"))
}

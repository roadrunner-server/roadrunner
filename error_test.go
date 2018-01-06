package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorkerError_Error(t *testing.T) {
	e := WorkerError("error")
	assert.Equal(t, "error", e.Error())
}

func TestJobError_Error(t *testing.T) {
	e := JobError([]byte("error"))
	assert.Equal(t, "error", e.Error())
}

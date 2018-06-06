package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"errors"
)

func Test_JobError_Error(t *testing.T) {
	e := JobError([]byte("error"))
	assert.Equal(t, "error", e.Error())
}

func Test_WorkerError_Error(t *testing.T) {
	e := WorkerError{Worker: nil, Caused: errors.New("error")}
	assert.Equal(t, "error", e.Error())
}

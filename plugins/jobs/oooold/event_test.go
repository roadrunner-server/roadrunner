package oooold

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJobEvent_Elapsed(t *testing.T) {
	e := &JobEvent{
		ID:      "id",
		Job:     &Job{},
		start:   time.Now(),
		elapsed: time.Millisecond,
	}

	assert.Equal(t, time.Millisecond, e.Elapsed())
}

func TestJobError_Elapsed(t *testing.T) {
	e := &JobError{
		ID:      "id",
		Job:     &Job{},
		start:   time.Now(),
		elapsed: time.Millisecond,
	}

	assert.Equal(t, time.Millisecond, e.Elapsed())
}

func TestJobError_Error(t *testing.T) {
	e := &JobError{
		ID:      "id",
		Job:     &Job{},
		start:   time.Now(),
		elapsed: time.Millisecond,
		Caused:  errors.New("error"),
	}

	assert.Equal(t, time.Millisecond, e.Elapsed())
	assert.Equal(t, "error", e.Error())
}

func TestPipelineError_Error(t *testing.T) {
	e := &PipelineError{
		Pipeline: &Pipeline{},
		Caused:   errors.New("error"),
	}

	assert.Equal(t, "error", e.Error())
}

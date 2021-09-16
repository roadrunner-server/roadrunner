package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewState(t *testing.T) {
	st := NewWorkerState(StateErrored)

	assert.Equal(t, "errored", st.String())

	assert.Equal(t, "inactive", NewWorkerState(StateInactive).String())
	assert.Equal(t, "ready", NewWorkerState(StateReady).String())
	assert.Equal(t, "working", NewWorkerState(StateWorking).String())
	assert.Equal(t, "stopped", NewWorkerState(StateStopped).String())
	assert.Equal(t, "undefined", NewWorkerState(1000).String())
}

func Test_IsActive(t *testing.T) {
	assert.False(t, NewWorkerState(StateInactive).IsActive())
	assert.True(t, NewWorkerState(StateReady).IsActive())
	assert.True(t, NewWorkerState(StateWorking).IsActive())
	assert.False(t, NewWorkerState(StateStopped).IsActive())
	assert.False(t, NewWorkerState(StateErrored).IsActive())
}

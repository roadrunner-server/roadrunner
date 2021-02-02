package internal

import (
	"testing"

	"github.com/spiral/roadrunner/v2/pkg/states"
	"github.com/stretchr/testify/assert"
)

func Test_NewState(t *testing.T) {
	st := NewWorkerState(states.StateErrored)

	assert.Equal(t, "errored", st.String())

	assert.Equal(t, "inactive", NewWorkerState(states.StateInactive).String())
	assert.Equal(t, "ready", NewWorkerState(states.StateReady).String())
	assert.Equal(t, "working", NewWorkerState(states.StateWorking).String())
	assert.Equal(t, "stopped", NewWorkerState(states.StateStopped).String())
	assert.Equal(t, "undefined", NewWorkerState(1000).String())
}

func Test_IsActive(t *testing.T) {
	assert.False(t, NewWorkerState(states.StateInactive).IsActive())
	assert.True(t, NewWorkerState(states.StateReady).IsActive())
	assert.True(t, NewWorkerState(states.StateWorking).IsActive())
	assert.False(t, NewWorkerState(states.StateStopped).IsActive())
	assert.False(t, NewWorkerState(states.StateErrored).IsActive())
}

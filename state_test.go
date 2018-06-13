package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewState(t *testing.T) {
	st := newState(StateErrored)

	assert.Equal(t, "errored", st.String())

	assert.Equal(t, "inactive", newState(StateInactive).String())
	assert.Equal(t, "ready", newState(StateReady).String())
	assert.Equal(t, "working", newState(StateWorking).String())
	assert.Equal(t, "stopped", newState(StateStopped).String())
	assert.Equal(t, "undefined", newState(1000).String())
}

func Test_IsActive(t *testing.T) {
	assert.False(t, newState(StateInactive).IsActive())
	assert.True(t, newState(StateReady).IsActive())
	assert.True(t, newState(StateWorking).IsActive())
	assert.False(t, newState(StateStopped).IsActive())
	assert.False(t, newState(StateErrored).IsActive())
}

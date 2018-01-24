package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NewState(t *testing.T) {
	st := newState(StateErrored)

	assert.Equal(t, "errored", st.String())
	assert.NotEqual(t, 0, st.Updated().Unix())
}

func Test_IsActive(t *testing.T) {
	assert.False(t, newState(StateInactive).IsActive())
	assert.True(t, newState(StateReady).IsActive())
	assert.True(t, newState(StateWorking).IsActive())
	assert.False(t, newState(StateStopped).IsActive())
	assert.False(t, newState(StateErrored).IsActive())
}

package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewState(t *testing.T) {
	st := newState(StateAttached)

	assert.Equal(t, "attached", st.String())
	assert.NotEqual(t, 0, st.Updated().Unix())
}

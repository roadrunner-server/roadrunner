package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CancellerNoListeners(t *testing.T) {
	c := &canceller{}

	assert.NoError(t, c.cancel(1))
}

func Test_CancellerListenerError(t *testing.T) {
	c := &canceller{}
	c.register(1, func() error {
		return errors.New("failed")
	})

	assert.Error(t, c.cancel(1))
}

func Test_CancellerListenerDiscarded(t *testing.T) {
	c := &canceller{}
	c.register(1, func() error {
		return errors.New("failed")
	})

	c.discard(1)
	assert.NoError(t, c.cancel(1))
}

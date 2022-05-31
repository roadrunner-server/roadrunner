package stop_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/stop"

	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	cmd := stop.NewCommand(toPtr(false), toPtr(false))

	assert.Equal(t, "stop", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestCommandTrue(t *testing.T) {
	cmd := stop.NewCommand(toPtr(true), toPtr(true))

	assert.Equal(t, "stop", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func toPtr[T any](val T) *T {
	return &val
}

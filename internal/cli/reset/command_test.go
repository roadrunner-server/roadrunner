package reset_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/reset"

	"github.com/roadrunner-server/config/v2"
	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	cmd := reset.NewCommand(&config.Plugin{})

	assert.Equal(t, "reset", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestExecution(t *testing.T) {
	t.Skip("Command execution is not implemented yet")
}

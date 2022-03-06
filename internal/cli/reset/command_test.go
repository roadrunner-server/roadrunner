package reset_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/reset"

	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	path := ""
	cmd := reset.NewCommand(&path)

	assert.Equal(t, "reset", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestExecution(t *testing.T) {
	t.Skip("Command execution is not implemented yet")
}

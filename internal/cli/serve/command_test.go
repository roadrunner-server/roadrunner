package serve_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/serve"

	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	path := ""
	cmd := serve.NewCommand(nil, &path, nil)

	assert.Equal(t, "serve", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestCommandNil(t *testing.T) {
	cmd := serve.NewCommand(nil, nil, nil)

	assert.Equal(t, "serve", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestExecution(t *testing.T) {
	t.Skip("Command execution is not implemented yet")
}

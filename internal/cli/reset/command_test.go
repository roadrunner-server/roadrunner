package reset_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/reset"

	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	path := ""
	f := false
	cmd := reset.NewCommand(&path, nil, &f)

	assert.Equal(t, "reset", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

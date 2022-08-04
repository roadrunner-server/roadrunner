package jobs_test

import (
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/jobs"
	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	path := ""
	f := false
	cmd := jobs.NewCommand(&path, nil, &f)

	assert.Equal(t, "jobs", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

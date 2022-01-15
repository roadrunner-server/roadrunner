package workers_test

import (
	"testing"

	"github.com/spiral/roadrunner-binary/v2/internal/cli/workers"

	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestCommandProperties(t *testing.T) {
	cmd := workers.NewCommand(&config.Plugin{})

	assert.Equal(t, "workers", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestCommandFlags(t *testing.T) {
	cmd := workers.NewCommand(&config.Plugin{})

	cases := []struct {
		giveName      string
		wantShorthand string
		wantDefault   string
	}{
		{giveName: "interactive", wantShorthand: "i", wantDefault: "false"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag [%s] was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
			assert.Equal(t, tt.wantDefault, flag.DefValue)
		})
	}
}

func TestExecution(t *testing.T) {
	t.Skip("Command execution is not implemented yet")
}

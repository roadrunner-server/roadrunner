package container_test

import (
	"testing"
	"time"

	"github.com/roadrunner-server/config/v3"
	"github.com/roadrunner-server/roadrunner/v2/container"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestNewConfig_SuccessfulReading(t *testing.T) {
	c, err := container.NewConfig("test/endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*10, c.GracePeriod)
	assert.True(t, c.PrintGraph)
	assert.Equal(t, slog.LevelWarn, c.LogLevel)
}

func TestNewConfig_WithoutEndureKey(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := container.NewConfig("test/without_endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*30, c.GracePeriod)
	assert.False(t, c.PrintGraph)
	assert.Equal(t, slog.LevelError, c.LogLevel)
}

func TestNewConfig_LoggingLevels(t *testing.T) {
	for _, tt := range []struct {
		path      string
		giveLevel string
		wantLevel slog.Leveler
		wantError bool
	}{
		{path: "test/endure_ok_debug.yaml", giveLevel: "debug", wantLevel: slog.LevelDebug},
		{path: "test/endure_ok_info.yaml", giveLevel: "info", wantLevel: slog.LevelInfo},
		{path: "test/endure_ok_warn.yaml", giveLevel: "warn", wantLevel: slog.LevelWarn},
		{path: "test/endure_ok_error.yaml", giveLevel: "error", wantLevel: slog.LevelError},

		{path: "test/endure_ok_foobar.yaml", giveLevel: "foobar", wantError: true},
	} {
		tt := tt
		t.Run(tt.giveLevel, func(t *testing.T) {
			cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("endure:\n  log_level: " + tt.giveLevel)}
			assert.NoError(t, cfgPlugin.Init())

			c, err := container.NewConfig(tt.path)

			if tt.wantError {
				assert.Nil(t, c)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown log level")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
				assert.Equal(t, tt.wantLevel, c.LogLevel)
			}
		})
	}
}

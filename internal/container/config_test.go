package container_test

import (
	"testing"
	"time"

	"github.com/spiral/roadrunner-binary/v2/internal/container"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig_SuccessfulReading(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte(`
endure:
  grace_period: 10s
  print_graph: true
  retry_on_fail: true
  log_level: warn
`)}
	assert.NoError(t, cfgPlugin.Init())

	c, err := container.NewConfig(cfgPlugin)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*10, c.GracePeriod)
	assert.True(t, c.PrintGraph)
	assert.True(t, c.RetryOnFail)
	assert.Equal(t, endure.WarnLevel, c.LogLevel)
}

func TestNewConfig_WithoutEndureKey(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := container.NewConfig(cfgPlugin)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*30, c.GracePeriod)
	assert.False(t, c.PrintGraph)
	assert.False(t, c.RetryOnFail)
	assert.Equal(t, endure.ErrorLevel, c.LogLevel)
}

func TestNewConfig_LoggingLevels(t *testing.T) {
	for _, tt := range []struct {
		giveLevel string
		wantLevel endure.Level
		wantError bool
	}{
		{giveLevel: "debug", wantLevel: endure.DebugLevel},
		{giveLevel: "info", wantLevel: endure.InfoLevel},
		{giveLevel: "warn", wantLevel: endure.WarnLevel},
		{giveLevel: "warning", wantLevel: endure.WarnLevel},
		{giveLevel: "error", wantLevel: endure.ErrorLevel},
		{giveLevel: "panic", wantLevel: endure.PanicLevel},
		{giveLevel: "fatal", wantLevel: endure.FatalLevel},

		{giveLevel: "foobar", wantError: true},
	} {
		tt := tt
		t.Run(tt.giveLevel, func(t *testing.T) {
			cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("endure:\n  log_level: " + tt.giveLevel)}
			assert.NoError(t, cfgPlugin.Init())

			c, err := container.NewConfig(cfgPlugin)

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

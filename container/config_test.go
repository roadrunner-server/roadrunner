package container_test

import (
	"testing"
	"time"

	"github.com/roadrunner-server/config/v3"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/roadrunner/v2/container"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig_SuccessfulReading(t *testing.T) {
	c, err := container.NewConfig("test/endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*10, c.GracePeriod)
	assert.True(t, c.PrintGraph)
	assert.Equal(t, endure.WarnLevel, c.LogLevel)
}

func TestNewConfig_WithoutEndureKey(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := container.NewConfig("test/without_endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, time.Second*30, c.GracePeriod)
	assert.False(t, c.PrintGraph)
	assert.Equal(t, endure.ErrorLevel, c.LogLevel)
}

func TestNewConfig_LoggingLevels(t *testing.T) {
	for _, tt := range []struct {
		path      string
		giveLevel string
		wantLevel endure.Level
		wantError bool
	}{
		{path: "test/endure_ok_debug.yaml", giveLevel: "debug", wantLevel: endure.DebugLevel},
		{path: "test/endure_ok_info.yaml", giveLevel: "info", wantLevel: endure.InfoLevel},
		{path: "test/endure_ok_warn.yaml", giveLevel: "warn", wantLevel: endure.WarnLevel},
		{path: "test/endure_ok_panic.yaml", giveLevel: "panic", wantLevel: endure.PanicLevel},
		{path: "test/endure_ok_fatal.yaml", giveLevel: "fatal", wantLevel: endure.FatalLevel},

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

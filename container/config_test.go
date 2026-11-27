package container_test

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/roadrunner/v3/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_SuccessfulReading(t *testing.T) {
	c, err := container.NewConfig("test/endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	ll, err := container.ParseLogLevel(c.LogLevel)
	assert.NoError(t, err)

	assert.Equal(t, time.Second*10, c.GracePeriod)
	assert.True(t, c.PrintGraph)
	assert.Equal(t, slog.LevelWarn, ll.Level())
}

func TestNewConfig_WithoutEndureKey(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := container.NewConfig("test/without_endure_ok.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, c)

	ll, err := container.ParseLogLevel(c.LogLevel)
	assert.NoError(t, err)

	assert.Equal(t, time.Second*30, c.GracePeriod)
	assert.False(t, c.PrintGraph)
	assert.Equal(t, slog.LevelError, ll.Level())
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
		t.Run(tt.giveLevel, func(t *testing.T) {
			cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("endure:\n  log_level: " + tt.giveLevel)}
			assert.NoError(t, cfgPlugin.Init())

			c, err := container.NewConfig(tt.path)
			assert.NotNil(t, c)
			ll, err2 := container.ParseLogLevel(c.LogLevel)

			if tt.wantError {
				assert.Error(t, err2)
				assert.Contains(t, err2.Error(), "unknown log level")
			} else {
				assert.NoError(t, err)
				assert.NoError(t, err2)
				assert.Equal(t, tt.wantLevel, ll.Level())
			}
		})
	}
}

func TestNewConfig_Includes(t *testing.T) {
	tests := []struct {
		name            string
		mainBody        string
		includeBody     string
		flags           []string
		wantGracePeriod time.Duration
		wantLogLevel    string
		wantPrintGraph  bool
	}{
		{
			name:     "included grace_period applies when main omits endure",
			mainBody: "",
			includeBody: `version: "3"
endure:
  grace_period: 15s
`,
			wantGracePeriod: 15 * time.Second,
			wantLogLevel:    "error",
		},
		{
			name: "included grace_period overrides main endure",
			mainBody: `endure:
  grace_period: 45s
`,
			includeBody: `version: "3"
endure:
  grace_period: 15s
`,
			wantGracePeriod: 15 * time.Second,
			wantLogLevel:    "error",
		},
		{
			name: "split endure settings merge across main and include",
			mainBody: `endure:
  log_level: debug
  print_graph: true
`,
			includeBody: `version: "3"
endure:
  grace_period: 15s
`,
			wantGracePeriod: 15 * time.Second,
			wantLogLevel:    "debug",
			wantPrintGraph:  true,
		},
		{
			name:     "included env default expands for endure grace_period",
			mainBody: "",
			includeBody: `version: "3"
endure:
  grace_period: ${RR_TEST_ENDURE_GRACE:-12s}
`,
			wantGracePeriod: 12 * time.Second,
			wantLogLevel:    "error",
		},
		{
			name:            "CLI flag sets endure grace_period when main omits endure",
			mainBody:        "",
			flags:           []string{"endure.grace_period=20s"},
			wantGracePeriod: 20 * time.Second,
			wantLogLevel:    "error",
		},
		{
			name: "CLI flag overrides main endure grace_period",
			mainBody: `endure:
  grace_period: 45s
`,
			flags:           []string{"endure.grace_period=20s"},
			wantGracePeriod: 20 * time.Second,
			wantLogLevel:    "error",
		},
		{
			name: "single file control",
			mainBody: `endure:
  grace_period: 15s
  log_level: info
`,
			wantGracePeriod: 15 * time.Second,
			wantLogLevel:    "info",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mainContent := "version: \"3\"\n"
			if tc.includeBody != "" {
				incPath := filepath.Join(dir, "base.yaml")
				require.NoError(t, os.WriteFile(incPath, []byte(tc.includeBody), 0o600))
				mainContent += fmt.Sprintf("include:\n  - %q\n", incPath)
			}
			mainContent += tc.mainBody

			mainPath := filepath.Join(dir, "main.yaml")
			require.NoError(t, os.WriteFile(mainPath, []byte(mainContent), 0o600))

			cfg, err := container.NewConfig(mainPath, tc.flags...)
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.wantGracePeriod, cfg.GracePeriod)
			assert.Equal(t, tc.wantLogLevel, cfg.LogLevel)
			assert.Equal(t, tc.wantPrintGraph, cfg.PrintGraph)
		})
	}
}

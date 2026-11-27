package lib_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/roadrunner-server/informer/v6"
	"github.com/roadrunner-server/resetter/v6"
	"github.com/roadrunner-server/roadrunner/v3/lib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFailsOnMissingConfig(t *testing.T) {
	_, err := lib.NewRR("config/file/does/not/exist/.rr.yaml", []string{}, lib.DefaultPluginsList())
	assert.NotNil(t, err)
}

const testConfigWithVersion = `
version: '3'
server:
  command: "php src/index.php"
  relay:  "pipes"

endure:
  grace_period: 1s
`

const testConfig = `
server:
  command: "php src/index.php"
  relay:  "pipes"

endure:
  grace_period: 1s
`

func makeConfig(t *testing.T, configYaml string) string {
	cfgFile := os.TempDir() + "/.rr.yaml"
	err := os.WriteFile(cfgFile, []byte(configYaml), 0600)
	assert.NoError(t, err)

	return cfgFile
}

func TestNewWithOldConfig(t *testing.T) {
	cfgFile := makeConfig(t, testConfig)
	_, err := lib.NewRR(cfgFile, []string{}, lib.DefaultPluginsList())
	assert.Error(t, err)

	t.Cleanup(func() {
		_ = os.Remove(cfgFile)
	})
}

func TestNewWithConfig(t *testing.T) {
	cfgFile := makeConfig(t, testConfigWithVersion)
	rr, err := lib.NewRR(cfgFile, []string{}, lib.DefaultPluginsList())
	assert.NoError(t, err)
	assert.Equal(t, "3", rr.Version)

	t.Cleanup(func() {
		_ = os.Remove(cfgFile)
	})
}

func TestServeStop(t *testing.T) {
	cfgFile := makeConfig(t, testConfigWithVersion)
	plugins := []any{
		&informer.Plugin{},
		&resetter.Plugin{},
	}
	rr, err := lib.NewRR(cfgFile, []string{}, plugins)
	assert.NoError(t, err)

	errchan := make(chan error, 1)
	stopchan := make(chan struct{}, 1)

	go func() {
		errchan <- rr.Serve()
		stopchan <- struct{}{}
	}()

	rr.Stop()
	time.Sleep(time.Second * 2)

	assert.Equal(t, struct{}{}, <-stopchan)
	assert.Nil(t, <-errchan)

	t.Cleanup(func() {
		_ = os.Remove(cfgFile)
	})
}

type deadlineConfigurer interface {
	Get(name string) any
	GracefulTimeout() time.Duration
}

type deadlineRecorderPlugin struct {
	visibleGrace    any
	providerTimeout time.Duration
	stopBudget      time.Duration
	hasDeadline     bool
}

func (p *deadlineRecorderPlugin) Init(cfg deadlineConfigurer) error {
	p.visibleGrace = cfg.Get("endure.grace_period")
	p.providerTimeout = cfg.GracefulTimeout()
	return nil
}

func (p *deadlineRecorderPlugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *deadlineRecorderPlugin) Stop(ctx context.Context) error {
	dl, ok := ctx.Deadline()
	p.hasDeadline = ok
	if ok {
		p.stopBudget = time.Until(dl)
	}
	return nil
}

func (p *deadlineRecorderPlugin) Name() string {
	return "deadline_recorder"
}

func TestNewRR_IncludedEndureGracePeriod(t *testing.T) {
	tests := []struct {
		name          string
		mainBody      string
		includeBody   string
		override      []string
		wantGrace     time.Duration
		wantGraceText string
	}{
		{
			name:     "included grace_period configures Endure stop deadline",
			mainBody: "",
			includeBody: `version: "3"
endure:
  grace_period: 15s
`,
			wantGrace:     15 * time.Second,
			wantGraceText: "15s",
		},
		{
			name: "included grace_period overrides main file deadline",
			mainBody: `endure:
  grace_period: 45s
`,
			includeBody: `version: "3"
endure:
  grace_period: 15s
`,
			wantGrace:     15 * time.Second,
			wantGraceText: "15s",
		},
		{
			name: "CLI override configures Endure stop deadline",
			mainBody: `endure:
  grace_period: 45s
`,
			override:      []string{"endure.grace_period=20s"},
			wantGrace:     20 * time.Second,
			wantGraceText: "20s",
		},
		{
			name: "single file control",
			mainBody: `endure:
  grace_period: 15s
`,
			wantGrace:     15 * time.Second,
			wantGraceText: "15s",
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

			rec := &deadlineRecorderPlugin{}
			rr, err := lib.NewRR(mainPath, tc.override, []any{rec})
			require.NoError(t, err)

			serveErr := make(chan error, 1)
			go func() {
				serveErr <- rr.Serve()
			}()

			rr.Stop()
			require.NoError(t, <-serveErr)

			assert.Equal(t, tc.wantGraceText, rec.visibleGrace)
			assert.Equal(t, tc.wantGrace, rec.providerTimeout)
			require.True(t, rec.hasDeadline)
			assert.InDelta(t, tc.wantGrace.Seconds(), rec.stopBudget.Seconds(), 1.0)
		})
	}
}

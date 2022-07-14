package roadrunner_test

import (
	"os"
	"testing"
	"time"

	"github.com/roadrunner-server/endure/pkg/fsm"
	"github.com/roadrunner-server/informer/v2"
	"github.com/roadrunner-server/resetter/v2"
	"github.com/roadrunner-server/roadrunner/v2/roadrunner"
	"github.com/stretchr/testify/assert"
)

func TestNewFailsOnMissingConfig(t *testing.T) {
	_, err := roadrunner.NewRR("config/file/does/not/exist/.rr.yaml", &[]string{}, roadrunner.DefaultPluginsList())
	assert.NotNil(t, err)
}

const testConfig = `
server:
  command: "php src/index.php"
  relay:  "unix://rr.sock"

endure:
  grace_period: 1s
`

func makeConfig(t *testing.T, configYaml string) string {
	cfgFile := os.TempDir() + "/.rr.yaml"
	err := os.WriteFile(cfgFile, []byte(configYaml), 0600)
	assert.Nil(t, err)

	return cfgFile
}

func TestNewWithConfig(t *testing.T) {
	cfgFile := makeConfig(t, testConfig)
	rr, err := roadrunner.NewRR(cfgFile, &[]string{}, roadrunner.DefaultPluginsList())
	assert.Nil(t, err)

	assert.Equal(t, "development", rr.BuildTime)
	assert.Equal(t, "local", rr.Version)
	assert.Equal(t, fsm.Initialized, rr.CurrentState())
}

func TestServeStop(t *testing.T) {
	cfgFile := makeConfig(t, testConfig)
	plugins := []interface{}{
		&informer.Plugin{},
		&resetter.Plugin{},
	}
	rr, err := roadrunner.NewRR(cfgFile, &[]string{}, plugins)
	assert.Nil(t, err)

	errchan := make(chan error, 1)
	stopchan := make(chan struct{}, 1)

	go func() {
		errchan <- rr.Serve()
		stopchan <- struct{}{}
	}()

	assert.Equal(t, rr.CurrentState(), fsm.Initialized)

	for rr.CurrentState() != fsm.Started {
		time.Sleep(20 * time.Millisecond)
	}

	assert.Nil(t, serveError)
	assert.False(t, stopped)

	err = rr.Stop()
	assert.Nil(t, err)
	assert.Equal(t, fsm.Stopped, rr.CurrentState())
	assert.Equal(t, struct{}, <-stopped)
	assert.Nil(t, <-serveError)
}

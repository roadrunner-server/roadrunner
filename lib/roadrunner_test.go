package lib_test

import (
	"os"
	"testing"
	"time"

	"github.com/roadrunner-server/endure/pkg/fsm"
	"github.com/roadrunner-server/informer/v3"
	"github.com/roadrunner-server/resetter/v3"
	"github.com/roadrunner-server/roadrunner/v2/lib"
	"github.com/stretchr/testify/assert"
)

func TestNewFailsOnMissingConfig(t *testing.T) {
	_, err := lib.NewRR("config/file/does/not/exist/.rr.yaml", []string{}, lib.DefaultPluginsList())
	assert.NotNil(t, err)
}

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

func TestNewWithConfig(t *testing.T) {
	cfgFile := makeConfig(t, testConfig)
	rr, err := lib.NewRR(cfgFile, []string{}, lib.DefaultPluginsList())
	assert.NoError(t, err)

	assert.Equal(t, "2", string(rr.Version[0]))
	assert.Equal(t, fsm.Initialized, rr.CurrentState())

	t.Cleanup(func() {
		_ = os.Remove(cfgFile)
	})
}

func TestServeStop(t *testing.T) {
	cfgFile := makeConfig(t, testConfig)
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

	assert.Equal(t, rr.CurrentState(), fsm.Initialized)

	for rr.CurrentState() != fsm.Started {
		time.Sleep(20 * time.Millisecond)
	}

	rr.Stop()
	time.Sleep(time.Second * 2)

	assert.Equal(t, fsm.Stopped, rr.CurrentState())
	assert.Equal(t, struct{}{}, <-stopchan)
	assert.Nil(t, <-errchan)

	t.Cleanup(func() {
		_ = os.Remove(cfgFile)
	})
}

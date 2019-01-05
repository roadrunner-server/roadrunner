package util

import (
	"github.com/spiral/roadrunner"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

func TestServerState(t *testing.T) {
	rr := roadrunner.NewServer(
		&roadrunner.ServerConfig{
			Command:      "php ../tests/client.php echo tcp",
			Relay:        "tcp://:9007",
			RelayTimeout: 10 * time.Second,
			Pool: &roadrunner.Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())

	state, err := ServerState(rr)
	assert.NoError(t, err)

	assert.Len(t, state, runtime.NumCPU())
}

func TestServerState_Err(t *testing.T) {
	_, err := ServerState(nil)
	assert.Error(t, err)
}

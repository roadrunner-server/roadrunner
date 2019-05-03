package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestServer_PipesEcho(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func TestServer_NoPool(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestServer_SocketEcho(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command:      "php tests/client.php echo tcp",
			Relay:        "tcp://:9007",
			RelayTimeout: 10 * time.Second,
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func TestServer_Configure_BeforeStart(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	err := rr.Reconfigure(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      2,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	assert.NoError(t, err)

	assert.NoError(t, rr.Start())

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
	assert.Len(t, rr.Workers(), 2)
}

func TestServer_Stop_NotStarted(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})

	rr.Stop()
	assert.Nil(t, rr.Workers())
}

func TestServer_Reconfigure(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())
	assert.Len(t, rr.Workers(), 1)

	err := rr.Reconfigure(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      2,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	assert.NoError(t, err)

	assert.Len(t, rr.Workers(), 2)
}

func TestServer_Reset(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())
	assert.Len(t, rr.Workers(), 1)

	pid := *rr.Workers()[0].Pid
	assert.NoError(t, rr.Reset())
	assert.Len(t, rr.Workers(), 1)
	assert.NotEqual(t, pid, rr.Workers()[0].Pid)
}

func TestServer_ReplacePool(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	assert.NoError(t, rr.Start())

	constructed := make(chan interface{})
	rr.Listen(func(e int, ctx interface{}) {
		if e == EventPoolConstruct {
			close(constructed)
		}
	})

	rr.Reset()
	<-constructed

	for _, w := range rr.Workers() {
		assert.Equal(t, StateReady, w.state.Value())
	}
}

func TestServer_ServerFailure(t *testing.T) {
	rr := NewServer(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	defer rr.Stop()

	assert.NoError(t, rr.Start())

	failure := make(chan interface{})
	rr.Listen(func(e int, ctx interface{}) {
		if e == EventServerFailure {
			failure <- nil
		}
	})

	// emulating potential server failure
	rr.cfg.Command = "php tests/client.php echo broken-connection"
	rr.pool.(*StaticPool).cmd = func() *exec.Cmd {
		return exec.Command("php", "tests/client.php", "echo", "broken-connection")
	}

	// killing random worker and expecting pool to replace it
	rr.Workers()[0].cmd.Process.Kill()

	<-failure
	assert.True(t, true)
}

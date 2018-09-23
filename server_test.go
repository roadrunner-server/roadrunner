package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestServer_PipesEcho(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	assert.NoError(t, srv.Start())

	res, err := srv.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func TestServer_SocketEcho(t *testing.T) {
	srv := NewServer(
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
	defer srv.Stop()

	assert.NoError(t, srv.Start())

	res, err := srv.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func TestServer_Configure_BeforeStart(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	err := srv.Reconfigure(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      2,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	assert.NoError(t, err)

	assert.NoError(t, srv.Start())

	res, err := srv.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
	assert.Len(t, srv.Workers(), 2)
}

func TestServer_Stop_NotStarted(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})

	srv.Stop()
	assert.Nil(t, srv.Workers())
}

func TestServer_Reconfigure(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	assert.NoError(t, srv.Start())
	assert.Len(t, srv.Workers(), 1)

	err := srv.Reconfigure(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      2,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	assert.NoError(t, err)

	assert.Len(t, srv.Workers(), 2)
}

func TestServer_Reset(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	assert.NoError(t, srv.Start())
	assert.Len(t, srv.Workers(), 1)

	pid := *srv.Workers()[0].Pid
	assert.NoError(t, srv.Reset())
	assert.Len(t, srv.Workers(), 1)
	assert.NotEqual(t, pid, srv.Workers()[0].Pid)
}

func TestServer_ReplacePool(t *testing.T) {
	srv := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	assert.NoError(t, srv.Start())

	constructed := make(chan interface{})
	srv.Listen(func(e int, ctx interface{}) {
		if e == EventPoolConstruct {
			close(constructed)
		}
	})

	srv.Reset()
	<-constructed

	for _, w := range srv.Workers() {
		assert.Equal(t, StateReady, w.state.Value())
	}
}

func TestServer_ServerFailure(t *testing.T) {
	srv := NewServer(&ServerConfig{
		Command: "php tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	})
	defer srv.Stop()

	assert.NoError(t, srv.Start())

	failure := make(chan interface{})
	srv.Listen(func(e int, ctx interface{}) {
		if e == EventServerFailure {
			failure <- nil
		}
	})

	// emulating potential server failure
	srv.cfg.Command = "php tests/client.php echo broken-connection"
	srv.pool.(*StaticPool).cmd = func() *exec.Cmd {
		return exec.Command("php", "tests/client.php", "echo", "broken-connection")
	}

	// killing random worker and expecting pool to replace it
	srv.Workers()[0].cmd.Process.Kill()

	<-failure
	assert.True(t, true)
}

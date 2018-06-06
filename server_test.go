package roadrunner

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"runtime"
	"time"
	"os/exec"
)

func TestServer_PipesEcho(t *testing.T) {
	srv := NewServer(
		func() *exec.Cmd { return exec.Command("php", "php-src/tests/client.php", "echo", "pipes") },
		&ServerConfig{
			Relay: "pipes",
			Pool: Config{
				NumWorkers:      uint64(runtime.NumCPU()),
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
		func() *exec.Cmd { return exec.Command("php", "php-src/tests/client.php", "echo", "tcp") },
		&ServerConfig{
			Relay:        "tcp://:9007",
			RelayTimeout: 10 * time.Second,
			Pool: Config{
				NumWorkers:      uint64(runtime.NumCPU()),
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
		func() *exec.Cmd { return exec.Command("php", "php-src/tests/client.php", "echo", "pipes") },
		&ServerConfig{
			Relay: "pipes",
			Pool: Config{
				NumWorkers:      uint64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	err := srv.Reconfigure(&ServerConfig{
		Relay: "pipes",
		Pool: Config{
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

func TestServer_Reconfigure(t *testing.T) {
	srv := NewServer(
		func() *exec.Cmd { return exec.Command("php", "php-src/tests/client.php", "echo", "pipes") },
		&ServerConfig{
			Relay: "pipes",
			Pool: Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer srv.Stop()

	assert.NoError(t, srv.Start())
	assert.Len(t, srv.Workers(), 1)

	err := srv.Reconfigure(&ServerConfig{
		Relay: "pipes",
		Pool: Config{
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
		func() *exec.Cmd { return exec.Command("php", "php-src/tests/client.php", "echo", "pipes") },
		&ServerConfig{
			Relay: "pipes",
			Pool: Config{
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

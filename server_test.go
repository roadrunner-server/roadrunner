package roadrunner

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"runtime"
	"time"
)

func TestServer_PipesEcho(t *testing.T) {
	srv := NewServer(&ServerConfig{
		Command: "php php-src/tests/client.php echo pipes",
		Relay:   "pipes",
		Pool: &Config{
			NumWorkers:      uint64(runtime.NumCPU()),
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	}, nil)
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
	srv := NewServer(&ServerConfig{
		Command:      "php php-src/tests/client.php echo tcp",
		Relay:        "tcp://:9007",
		RelayTimeout: 10 * time.Second,
		Pool: &Config{
			NumWorkers:      uint64(runtime.NumCPU()),
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	}, nil)
	defer srv.Stop()

	assert.NoError(t, srv.Start())

	res, err := srv.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

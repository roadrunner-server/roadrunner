package http

import (
	"encoding/json"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type mockCfg struct{ cfg string }

func (cfg *mockCfg) Get(name string) service.Config  { return nil }
func (cfg *mockCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func Test_Config_Hydrate_Error1(t *testing.T) {
	cfg := &mockCfg{`{"enable": true}`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{"dir": "/dir/"`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Valid(t *testing.T) {
	cfg := &Config{
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.NoError(t, cfg.Valid())
}

func Test_Config_Valid_SSL(t *testing.T) {
	cfg := &Config{
		Address: ":8080",
		SSL: SSLConfig{
			Cert: "fixtures/server.crt",
			Key:  "fixtures/server.key",
		},
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Hydrate(&testCfg{httpCfg: "{}"}))

	assert.NoError(t, cfg.Valid())
	assert.True(t, cfg.EnableTLS())
	assert.Equal(t, 443, cfg.SSL.Port)
}

func Test_Config_SSL_No_key(t *testing.T) {
	cfg := &Config{
		Address: ":8080",
		SSL: SSLConfig{
			Cert: "fixtures/server.crt",
		},
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_SSL_No_Cert(t *testing.T) {
	cfg := &Config{
		Address: ":8080",
		SSL: SSLConfig{
			Key: "fixtures/server.key",
		},
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_NoUploads(t *testing.T) {
	cfg := &Config{
		Address:    ":8080",
		MaxRequest: 1024,
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_NoWorkers(t *testing.T) {
	cfg := &Config{
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_NoPool(t *testing.T) {
	cfg := &Config{
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      0,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_DeadPool(t *testing.T) {
	cfg := &Config{
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_InvalidAddress(t *testing.T) {
	cfg := &Config{
		Address:    "",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		},
	}

	assert.Error(t, cfg.Valid())
}

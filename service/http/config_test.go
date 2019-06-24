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
	cfg := &mockCfg{`{"address": "localhost:8080"}`}
	c := &Config{}

	assert.NoError(t, c.Hydrate(cfg))
}

func Test_Config_Hydrate_Error2(t *testing.T) {
	cfg := &mockCfg{`{"dir": "/dir/"`}
	c := &Config{}

	assert.Error(t, c.Hydrate(cfg))
}

func Test_Config_Valid(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
		HTTP2: &HTTP2Config{
			Enabled: true,
		},
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

func Test_Trusted_Subnets(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
		},
		TrustedSubnets: []string{"200.1.0.0/16"},
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

	assert.NoError(t, cfg.parseCIDRs())

	assert.True(t, cfg.IsTrusted("200.1.0.10"))
	assert.False(t, cfg.IsTrusted("127.0.0.0.1"))
}

func Test_Trusted_Subnets_Err(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
		},
		TrustedSubnets: []string{"200.1.0.0"},
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

	assert.Error(t, cfg.parseCIDRs())
}

func Test_Config_Valid_SSL(t *testing.T) {
	cfg := &Config{
		Address: ":8080",
		SSL: SSLConfig{
			Cert: "fixtures/server.crt",
			Key:  "fixtures/server.key",
		},
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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
		Address:        ":8080",
		MaxRequestSize: 1024,
		HTTP2: &HTTP2Config{
			Enabled: true,
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

func Test_Config_NoHTTP2(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
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

func Test_Config_NoWorkers(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
		HTTP2: &HTTP2Config{
			Enabled: true,
		},
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Config_NoPool(t *testing.T) {
	cfg := &Config{
		Address:        ":8080",
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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
		Address:        ":8080",
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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
		Address:        "unexpected_address",
		MaxRequestSize: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		HTTP2: &HTTP2Config{
			Enabled: true,
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

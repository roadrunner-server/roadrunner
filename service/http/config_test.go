package http

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
	"github.com/spiral/roadrunner"
	"time"
)

func Test_Config_Valid(t *testing.T) {
	cfg := &Config{
		Enable:     true,
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php php-src/tests/client.php echo pipes",
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

func Test_Config_NoUploads(t *testing.T) {
	cfg := &Config{
		Enable:     true,
		Address:    ":8080",
		MaxRequest: 1024,
		Workers: &roadrunner.ServerConfig{
			Command: "php php-src/tests/client.php echo pipes",
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
		Enable:     true,
		Address:    ":8080",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
	}

	assert.Error(t, cfg.Valid())
}

func Test_Confi_InvalidAddress(t *testing.T) {
	cfg := &Config{
		Enable:     true,
		Address:    "",
		MaxRequest: 1024,
		Uploads: &UploadsConfig{
			Dir:    os.TempDir(),
			Forbid: []string{".go"},
		},
		Workers: &roadrunner.ServerConfig{
			Command: "php php-src/tests/client.php echo pipes",
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

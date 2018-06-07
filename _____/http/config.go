package http

import (
	"fmt"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/_____/utils"
	"os"
	"path"
	"strings"
)

// Configures RoadRunner HTTP server.
type Config struct {
	// serve enables static file serving from desired root directory.
	ServeStatic bool

	// Root directory, required when serve set to true.
	Root string

	// TmpDir contains name of temporary directory to store uploaded files passed to underlying PHP process.
	TmpDir string

	// MaxRequest specified max size for payload body in bytes, set 0 to unlimited.
	MaxRequest int64

	// ForbidUploads specifies list of file extensions which are forbidden for uploads.
	// Example: .php, .exe, .bat, .htaccess and etc.
	ForbidUploads []string
}

// ForbidUploads must return true if file extension is not allowed for the upload.
func (cfg Config) Forbidden(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range cfg.ForbidUploads {
		if ext == v {
			return true
		}
	}

	return false
}

type serviceConfig struct {
	Enabled    bool
	Host       string
	Port       string
	MaxRequest string
	Static     struct {
		Serve bool
		Root  string
	}

	Uploads struct {
		TmpDir string
		Forbid []string
	}

	Pool service.PoolConfig

	//todo: verbose ?
}

func (cfg *serviceConfig) httpAddr() string {
	return fmt.Sprintf("%s:%v", cfg.Host, cfg.Port)
}

func (cfg *serviceConfig) httpConfig() *Config {
	tmpDir := cfg.Uploads.TmpDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}

	return &Config{
		ServeStatic:   cfg.Static.Serve,
		Root:          cfg.Static.Root,
		TmpDir:        tmpDir,
		MaxRequest:    utils.ParseSize(cfg.MaxRequest),
		ForbidUploads: cfg.Uploads.Forbid,
	}
}

func (cfg *serviceConfig) Valid() error {
	return nil
}

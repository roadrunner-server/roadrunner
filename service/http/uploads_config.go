package http

import (
	"os"
	"path"
	"strings"
)

// UploadsConfig describes file location and controls access to them.
type UploadsConfig struct {
	// Dir contains name of directory to control access to.
	Dir string

	// Forbid specifies list of file extensions which are forbidden for access.
	// Example: .php, .exe, .bat, .htaccess and etc.
	Forbid []string
}

// InitDefaults sets missing values to their default values.
func (cfg *UploadsConfig) InitDefaults() error {
	cfg.Forbid = []string{".php", ".exe", ".bat"}
	return nil
}

// TmpDir returns temporary directory.
func (cfg *UploadsConfig) TmpDir() string {
	if cfg.Dir != "" {
		return cfg.Dir
	}

	return os.TempDir()
}

// Forbids must return true if file extension is not allowed for the upload.
func (cfg *UploadsConfig) Forbids(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range cfg.Forbid {
		if ext == v {
			return true
		}
	}

	return false
}

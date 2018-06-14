package static

import (
	"github.com/pkg/errors"
	"os"
	"path"
	"strings"
)

// Config describes file location and controls access to them.
type Config struct {
	// Enables StaticFile service.
	Enable bool

	// Dir contains name of directory to control access to.
	Dir string

	// Forbid specifies list of file extensions which are forbidden for access.
	// Example: .php, .exe, .bat, .htaccess and etc.
	Forbid []string
}

// Forbid must return true if file extension is not allowed for the upload.
func (cfg *Config) Forbids(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range cfg.Forbid {
		if ext == v {
			return true
		}
	}

	return false
}

// Valid validates existence of directory.
func (cfg *Config) Valid() error {
	st, err := os.Stat(cfg.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("root directory does not exists")
		}

		return err
	}

	if !st.IsDir() {
		return errors.New("invalid root directory")
	}

	return nil
}

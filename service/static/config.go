package static

import (
	"fmt"
	"github.com/spiral/roadrunner/service"
	"os"
	"path"
	"strings"
)

// Config describes file location and controls access to them.
type Config struct {
	// Dir contains name of directory to control access to.
	Dir string

	// Forbid specifies list of file extensions which are forbidden for access.
	// Example: .php, .exe, .bat, .htaccess and etc.
	Forbid []string

	// Always specifies list of extensions which must always be served by static
	// service, even if file not found.
	Always []string
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Valid returns nil if config is valid.
func (c *Config) Valid() error {
	st, err := os.Stat(c.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("root directory '%s' does not exists", c.Dir)
		}

		return err
	}

	if !st.IsDir() {
		return fmt.Errorf("invalid root directory '%s'", c.Dir)
	}

	return nil
}

// AlwaysForbid must return true if file extension is not allowed for the upload.
func (c *Config) AlwaysForbid(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range c.Forbid {
		if ext == v {
			return true
		}
	}

	return false
}

// AlwaysServe must indicate that file is expected to be served by static service.
func (c *Config) AlwaysServe(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range c.Always {
		if ext == v {
			return true
		}
	}

	return false
}

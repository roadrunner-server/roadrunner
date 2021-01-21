package static

import (
	"os"
	"path"
	"strings"

	"github.com/spiral/errors"
)

// Config describes file location and controls access to them.
type Config struct {
	Static *struct {
		// Dir contains name of directory to control access to.
		Dir string

		// Forbid specifies list of file extensions which are forbidden for access.
		// Example: .php, .exe, .bat, .htaccess and etc.
		Forbid []string

		// Always specifies list of extensions which must always be served by static
		// service, even if file not found.
		Always []string

		// Request headers to add to every static.
		Request map[string]string

		// Response headers to add to every static.
		Response map[string]string
	}
}

// Valid returns nil if config is valid.
func (c *Config) Valid() error {
	const op = errors.Op("static_plugin_valid")
	st, err := os.Stat(c.Static.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.E(op, errors.Errorf("root directory '%s' does not exists", c.Static.Dir))
		}

		return err
	}

	if !st.IsDir() {
		return errors.E(op, errors.Errorf("invalid root directory '%s'", c.Static.Dir))
	}

	return nil
}

// AlwaysForbid must return true if file extension is not allowed for the upload.
func (c *Config) AlwaysForbid(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range c.Static.Forbid {
		if ext == v {
			return true
		}
	}

	return false
}

// AlwaysServe must indicate that file is expected to be served by static service.
func (c *Config) AlwaysServe(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range c.Static.Always {
		if ext == v {
			return true
		}
	}

	return false
}

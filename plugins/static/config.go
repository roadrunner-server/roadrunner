package static

import (
	"os"

	"github.com/spiral/errors"
)

// Config describes file location and controls access to them.
type Config struct {
	Static *struct {
		// Dir contains name of directory to control access to.
		// Default - "."
		Dir string

		// CalculateEtag can be true/false and used to calculate etag for the static
		CalculateEtag bool `mapstructure:"calculate_etag"`

		// Weak etag `W/`
		Weak bool

		// forbid specifies list of file extensions which are forbidden for access.
		// example: .php, .exe, .bat, .htaccess and etc.
		Forbid []string

		// Allow specifies list of file extensions which are allowed for access.
		// example: .php, .exe, .bat, .htaccess and etc.
		Allow []string

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

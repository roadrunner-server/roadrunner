package reload

import (
	"errors"
	"time"

	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
)

// Config is a Reload configuration point.
type Config struct {
	// Interval is a global refresh interval
	Interval time.Duration

	// Patterns is a global file patterns to watch. It will be applied to every directory in project
	Patterns []string

	// Services is set of services which would be reloaded in case of FS changes
	Services map[string]ServiceConfig
}

type ServiceConfig struct {
	// Enabled indicates that service must be watched, doest not required when any other option specified
	Enabled bool

	// Recursive is options to use nested files from root folder
	Recursive bool

	// Patterns is per-service specific files to watch
	Patterns []string

	// Dirs is per-service specific dirs which will be combined with Patterns
	Dirs []string

	// Ignore is set of files which would not be watched
	Ignore []string

	// service is a link to service to restart
	service *roadrunner.Controllable
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return nil
}

// InitDefaults sets missing values to their default values.
func (c *Config) InitDefaults() error {
	c.Interval = time.Second
	c.Patterns = []string{".php"}

	return nil
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	if c.Interval < time.Second {
		return errors.New("too short interval")
	}

	if c.Services == nil {
		return errors.New("should add at least 1 service")
	} else if len(c.Services) == 0 {
		return errors.New("service initialized, however, no config added")
	}

	return nil
}

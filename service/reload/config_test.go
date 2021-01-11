package reload

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Config_Valid(t *testing.T) {
	services := make(map[string]ServiceConfig)
	services["test"] = ServiceConfig{
		Recursive: false,
		Patterns:  nil,
		Dirs:      nil,
		Ignore:    nil,
		service:   nil,
	}

	cfg := &Config{
		Interval: time.Second,
		Patterns: nil,
		Services: services,
	}
	assert.NoError(t, cfg.Valid())
}

func Test_Fake_ServiceConfig(t *testing.T) {
	services := make(map[string]ServiceConfig)
	cfg := &Config{
		Interval: time.Microsecond,
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

func Test_Interval(t *testing.T) {
	services := make(map[string]ServiceConfig)
	services["test"] = ServiceConfig{
		Enabled:   false,
		Recursive: false,
		Patterns:  nil,
		Dirs:      nil,
		Ignore:    nil,
		service:   nil,
	}

	cfg := &Config{
		Interval: time.Millisecond, // should crash here
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

func Test_NoServiceConfig(t *testing.T) {
	cfg := &Config{
		Interval: time.Second,
		Patterns: nil,
		Services: nil,
	}
	assert.Error(t, cfg.Valid())
}

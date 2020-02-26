package reload

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
	cfg := &Config{
		Interval: time.Millisecond,
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

func Test_NoServiceConfig(t *testing.T) {
	services := make(map[string]ServiceConfig)
	cfg := &Config{
		Interval: time.Millisecond,
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

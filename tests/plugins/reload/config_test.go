package reload

import (
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/plugins/reload"
	"github.com/stretchr/testify/assert"
)

func Test_Config_Valid(t *testing.T) {
	services := make(map[string]reload.ServiceConfig)
	services["test"] = reload.ServiceConfig{
		Recursive: false,
		Patterns:  nil,
		Dirs:      nil,
		Ignore:    nil,
	}

	cfg := &reload.Config{
		Interval: time.Second,
		Patterns: nil,
		Services: services,
	}
	assert.NoError(t, cfg.Valid())
}

func Test_Fake_ServiceConfig(t *testing.T) {
	services := make(map[string]reload.ServiceConfig)
	cfg := &reload.Config{
		Interval: time.Microsecond,
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

func Test_Interval(t *testing.T) {
	services := make(map[string]reload.ServiceConfig)
	services["test"] = reload.ServiceConfig{
		Enabled:   false,
		Recursive: false,
		Patterns:  nil,
		Dirs:      nil,
		Ignore:    nil,
	}

	cfg := &reload.Config{
		Interval: time.Millisecond, // should crash here
		Patterns: nil,
		Services: services,
	}
	assert.Error(t, cfg.Valid())
}

func Test_NoServiceConfig(t *testing.T) {
	cfg := &reload.Config{
		Interval: time.Second,
		Patterns: nil,
		Services: nil,
	}
	assert.Error(t, cfg.Valid())
}

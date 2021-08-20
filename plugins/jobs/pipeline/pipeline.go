package pipeline

import (
	json "github.com/json-iterator/go"
	"github.com/spiral/roadrunner/v2/utils"
)

// Pipeline defines pipeline options.
type Pipeline map[string]interface{}

const (
	priority string = "priority"
	driver   string = "driver"
	name     string = "name"
)

// With pipeline value
func (p *Pipeline) With(name string, value interface{}) {
	(*p)[name] = value
}

// Name returns pipeline name.
func (p Pipeline) Name() string {
	return p.String(name, "")
}

// Driver associated with the pipeline.
func (p Pipeline) Driver() string {
	return p.String(driver, "")
}

// Has checks if value presented in pipeline.
func (p Pipeline) Has(name string) bool {
	if _, ok := p[name]; ok {
		return true
	}

	return false
}

// String must return option value as string or return default value.
func (p Pipeline) String(name string, d string) string {
	if value, ok := p[name]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}

	return d
}

// Int must return option value as string or return default value.
func (p Pipeline) Int(name string, d int) int {
	if value, ok := p[name]; ok {
		if i, ok := value.(int); ok {
			return i
		}
	}

	return d
}

// Bool must return option value as bool or return default value.
func (p Pipeline) Bool(name string, d bool) bool {
	if value, ok := p[name]; ok {
		if i, ok := value.(bool); ok {
			return i
		}
	}

	return d
}

// Map must return nested map value or empty config.
// Here might be sqs attributes or tags for example
func (p Pipeline) Map(name string, out map[string]string) error {
	if value, ok := p[name]; ok {
		if m, ok := value.(string); ok {
			err := json.Unmarshal(utils.AsBytes(m), &out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Priority returns default pipeline priority
func (p Pipeline) Priority() int64 {
	if value, ok := p[priority]; ok {
		if v, ok := value.(int64); ok {
			return v
		}
	}

	return 10
}

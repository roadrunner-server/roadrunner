package pipeline

import (
	"time"

	"github.com/spiral/errors"
)

// Pipelines is list of Pipeline.

type Pipelines []*Pipeline

func InitPipelines(pipes map[string]*Pipeline) (Pipelines, error) {
	const op = errors.Op("pipeline_init")
	out := make(Pipelines, 0)

	for name, pipe := range pipes {
		if pipe.Broker() == "" {
			return nil, errors.E(op, errors.Errorf("found the pipeline without defined broker"))
		}

		p := pipe.With("name", name)
		out = append(out, &p)
	}

	return out, nil
}

// Reverse returns pipelines in reversed order.
func (ps Pipelines) Reverse() Pipelines {
	out := make(Pipelines, len(ps))

	for i, p := range ps {
		out[len(ps)-i-1] = p
	}

	return out
}

// Broker return pipelines associated with specific broker.
func (ps Pipelines) Broker(broker string) Pipelines {
	out := make(Pipelines, 0)

	for _, p := range ps {
		if p.Broker() != broker {
			continue
		}

		out = append(out, p)
	}

	return out
}

// Names returns only pipelines with specified names.
func (ps Pipelines) Names(only ...string) Pipelines {
	out := make(Pipelines, 0)

	for _, name := range only {
		for _, p := range ps {
			if p.Name() == name {
				out = append(out, p)
			}
		}
	}

	return out
}

// Get returns pipeline by it'svc name.
func (ps Pipelines) Get(name string) *Pipeline {
	// possibly optimize
	for _, p := range ps {
		if p.Name() == name {
			return p
		}
	}

	return nil
}

// Pipeline defines pipeline options.
type Pipeline map[string]interface{}

// With pipeline value. Immutable.
func (p Pipeline) With(name string, value interface{}) Pipeline {
	out := make(map[string]interface{})
	for k, v := range p {
		out[k] = v
	}
	out[name] = value

	return out
}

// Name returns pipeline name.
func (p Pipeline) Name() string {
	return p.String("name", "")
}

// Broker associated with the pipeline.
func (p Pipeline) Broker() string {
	return p.String("broker", "")
}

// Has checks if value presented in pipeline.
func (p Pipeline) Has(name string) bool {
	if _, ok := p[name]; ok {
		return true
	}

	return false
}

// Map must return nested map value or empty config.
func (p Pipeline) Map(name string) Pipeline {
	out := make(map[string]interface{})

	if value, ok := p[name]; ok {
		if m, ok := value.(map[string]interface{}); ok {
			for k, v := range m {
				out[k] = v
			}
		}
	}

	return out
}

// Bool must return option value as string or return default value.
func (p Pipeline) Bool(name string, d bool) bool {
	if value, ok := p[name]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}

	return d
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

// Integer must return option value as string or return default value.
func (p Pipeline) Integer(name string, d int) int {
	if value, ok := p[name]; ok {
		if str, ok := value.(int); ok {
			return str
		}
	}

	return d
}

// Duration must return option value as time.Duration (seconds) or return default value.
func (p Pipeline) Duration(name string, d time.Duration) time.Duration {
	if value, ok := p[name]; ok {
		if str, ok := value.(int); ok {
			return time.Second * time.Duration(str)
		}
	}

	return d
}

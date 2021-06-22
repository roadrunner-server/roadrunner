package pipeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipeline_Map(t *testing.T) {
	pipe := Pipeline{"options": map[string]interface{}{"ttl": 10}}

	assert.Equal(t, 10, pipe.Map("options").Integer("ttl", 0))
	assert.Equal(t, 0, pipe.Map("other").Integer("ttl", 0))
}

func TestPipeline_MapString(t *testing.T) {
	pipe := Pipeline{"options": map[string]interface{}{"alias": "default"}}

	assert.Equal(t, "default", pipe.Map("options").String("alias", ""))
	assert.Equal(t, "", pipe.Map("other").String("alias", ""))
}

func TestPipeline_Bool(t *testing.T) {
	pipe := Pipeline{"value": true}

	assert.Equal(t, true, pipe.Bool("value", false))
	assert.Equal(t, true, pipe.Bool("other", true))
}

func TestPipeline_String(t *testing.T) {
	pipe := Pipeline{"value": "value"}

	assert.Equal(t, "value", pipe.String("value", ""))
	assert.Equal(t, "value", pipe.String("other", "value"))
}

func TestPipeline_Integer(t *testing.T) {
	pipe := Pipeline{"value": 1}

	assert.Equal(t, 1, pipe.Integer("value", 0))
	assert.Equal(t, 1, pipe.Integer("other", 1))
}

func TestPipeline_Duration(t *testing.T) {
	pipe := Pipeline{"value": 1}

	assert.Equal(t, time.Second, pipe.Duration("value", 0))
	assert.Equal(t, time.Second, pipe.Duration("other", time.Second))
}

func TestPipeline_Has(t *testing.T) {
	pipe := Pipeline{"options": map[string]interface{}{"ttl": 10}}

	assert.Equal(t, true, pipe.Has("options"))
	assert.Equal(t, false, pipe.Has("other"))
}

func TestPipeline_FilterBroker(t *testing.T) {
	pipes := Pipelines{
		&Pipeline{"name": "first", "broker": "a"},
		&Pipeline{"name": "second", "broker": "a"},
		&Pipeline{"name": "third", "broker": "b"},
		&Pipeline{"name": "forth", "broker": "b"},
	}

	filtered := pipes.Names("first", "third")
	assert.True(t, len(filtered) == 2)

	assert.Equal(t, "a", filtered[0].Broker())
	assert.Equal(t, "b", filtered[1].Broker())

	filtered = pipes.Names("first", "third").Reverse()
	assert.True(t, len(filtered) == 2)

	assert.Equal(t, "a", filtered[1].Broker())
	assert.Equal(t, "b", filtered[0].Broker())

	filtered = pipes.Broker("a")
	assert.True(t, len(filtered) == 2)

	assert.Equal(t, "first", filtered[0].Name())
	assert.Equal(t, "second", filtered[1].Name())

	filtered = pipes.Broker("a").Reverse()
	assert.True(t, len(filtered) == 2)

	assert.Equal(t, "first", filtered[1].Name())
	assert.Equal(t, "second", filtered[0].Name())
}

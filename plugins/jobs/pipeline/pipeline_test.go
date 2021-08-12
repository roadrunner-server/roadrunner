package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline_String(t *testing.T) {
	pipe := Pipeline{"value": "value"}

	assert.Equal(t, "value", pipe.String("value", ""))
	assert.Equal(t, "value", pipe.String("other", "value"))
}

func TestPipeline_Has(t *testing.T) {
	pipe := Pipeline{"options": map[string]interface{}{"ttl": 10}}

	assert.Equal(t, true, pipe.Has("options"))
	assert.Equal(t, false, pipe.Has("other"))
}

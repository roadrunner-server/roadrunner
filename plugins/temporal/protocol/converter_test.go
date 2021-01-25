package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

func Test_Passthough(t *testing.T) {
	codec := NewDataConverter(converter.GetDefaultDataConverter())

	value, err := codec.ToPayloads("test")
	assert.NoError(t, err)

	out := &common.Payloads{}

	assert.Len(t, out.Payloads, 0)
	assert.NoError(t, codec.FromPayloads(value, &out))

	assert.Len(t, out.Payloads, 1)
}

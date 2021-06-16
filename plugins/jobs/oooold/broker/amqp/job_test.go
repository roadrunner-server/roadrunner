package amqp

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Unpack_Errors(t *testing.T) {
	_, _, _, err := unpack(amqp.Delivery{
		Headers: map[string]interface{}{},
	})
	assert.Error(t, err)

	_, _, _, err = unpack(amqp.Delivery{
		Headers: map[string]interface{}{
			"rr-id": "id",
		},
	})
	assert.Error(t, err)

	_, _, _, err = unpack(amqp.Delivery{
		Headers: map[string]interface{}{
			"rr-id":      "id",
			"rr-attempt": int64(0),
		},
	})
	assert.Error(t, err)
}

package workflow

import (
	"sync/atomic"
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
)

func Test_MessageQueueFlushError(t *testing.T) {
	var index uint64
	mq := newMessageQueue(func() uint64 {
		return atomic.AddUint64(&index, 1)
	})

	mq.pushError(1, &failure.Failure{})
	assert.Len(t, mq.queue, 1)

	mq.flush()
	assert.Len(t, mq.queue, 0)
	assert.Equal(t, uint64(0), index)
}

func Test_MessageQueueFlushResponse(t *testing.T) {
	var index uint64
	mq := newMessageQueue(func() uint64 {
		return atomic.AddUint64(&index, 1)
	})

	mq.pushResponse(1, &common.Payloads{})
	assert.Len(t, mq.queue, 1)

	mq.flush()
	assert.Len(t, mq.queue, 0)
	assert.Equal(t, uint64(0), index)
}

func Test_MessageQueueCommandID(t *testing.T) {
	var index uint64
	mq := newMessageQueue(func() uint64 {
		return atomic.AddUint64(&index, 1)
	})

	n, err := mq.pushCommand(protocol.StartWorkflow{}, &common.Payloads{})
	assert.Equal(t, n, index)

	assert.NoError(t, err)
	assert.Len(t, mq.queue, 1)

	mq.flush()
	assert.Len(t, mq.queue, 0)
}

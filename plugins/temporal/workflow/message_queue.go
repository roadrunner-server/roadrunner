package workflow

import (
	rrt "github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/failure/v1"
)

type messageQueue struct {
	seqID func() uint64
	queue []rrt.Message
}

func newMessageQueue(sedID func() uint64) *messageQueue {
	return &messageQueue{
		seqID: sedID,
		queue: make([]rrt.Message, 0, 5),
	}
}

func (mq *messageQueue) flush() {
	mq.queue = mq.queue[0:0]
}

func (mq *messageQueue) allocateMessage(
	cmd interface{},
	payloads *common.Payloads,
) (id uint64, msg rrt.Message, err error) {
	msg = rrt.Message{
		ID:       mq.seqID(),
		Command:  cmd,
		Payloads: payloads,
	}

	return msg.ID, msg, nil
}

func (mq *messageQueue) pushCommand(cmd interface{}, payloads *common.Payloads) (id uint64, err error) {
	id, msg, err := mq.allocateMessage(cmd, payloads)
	if err != nil {
		return 0, err
	}

	mq.queue = append(mq.queue, msg)

	return id, nil
}

func (mq *messageQueue) pushResponse(id uint64, payloads *common.Payloads) {
	mq.queue = append(mq.queue, rrt.Message{ID: id, Payloads: payloads})
}

func (mq *messageQueue) pushError(id uint64, failure *failure.Failure) {
	mq.queue = append(mq.queue, rrt.Message{ID: id, Failure: failure})
}

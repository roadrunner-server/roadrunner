package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	All string = "All"
)

func (j *JobConsumer) listen() {
	for {
		select {
		case <-j.pauseCh:
			return
		default:
			message, err := j.client.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
				QueueUrl:              j.outputQ.QueueUrl,
				MaxNumberOfMessages:   j.prefetch,
				AttributeNames:        []types.QueueAttributeName{types.QueueAttributeName(ApproximateReceiveCount)},
				MessageAttributeNames: []string{All},
				VisibilityTimeout:     j.visibilityTimeout,
				WaitTimeSeconds:       j.waitTime,
			})
			if err != nil {
				j.log.Error("receive message", "error", err)
				continue
			}

			for i := 0; i < len(message.Messages); i++ {
				m := message.Messages[i]
				item, attempt, err := j.unpack(&m)
				if err != nil {
					_, errD := j.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
						QueueUrl:      j.outputQ.QueueUrl,
						ReceiptHandle: m.ReceiptHandle,
					})
					if errD != nil {
						j.log.Error("message unpack, failed to delete the message from the queue", "error", err)
						continue
					}

					j.log.Error("message unpack", "error", err)
					continue
				}

				if item.Options.CanRetry(int64(attempt)) {
					j.pq.Insert(item)
					continue
				}

				_, errD := j.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
					QueueUrl:      j.outputQ.QueueUrl,
					ReceiptHandle: m.ReceiptHandle,
				})
				if errD != nil {
					j.log.Error("message unpack, failed to delete the message from the queue", "error", err)
					continue
				}
			}
		}
	}
}

package sqs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/smithy-go"
)

const (
	// All - get all message attribute names
	All string = "All"

	// NonExistentQueue AWS error code
	NonExistentQueue string = "AWS.SimpleQueueService.NonExistentQueue"
)

func (j *JobConsumer) listen(ctx context.Context) { //nolint:gocognit
	for {
		select {
		case <-j.pauseCh:
			j.log.Warn("sqs listener stopped")
			return
		default:
			message, err := j.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              j.queueURL,
				MaxNumberOfMessages:   j.prefetch,
				AttributeNames:        []types.QueueAttributeName{types.QueueAttributeName(ApproximateReceiveCount)},
				MessageAttributeNames: []string{All},
				VisibilityTimeout:     j.visibilityTimeout,
				WaitTimeSeconds:       j.waitTime,
			})

			if err != nil {
				if oErr, ok := (err).(*smithy.OperationError); ok {
					if rErr, ok := oErr.Err.(*http.ResponseError); ok {
						if apiErr, ok := rErr.Err.(*smithy.GenericAPIError); ok {
							// in case of NonExistentQueue - recreate the queue
							if apiErr.Code == NonExistentQueue {
								j.log.Error("receive message", "error code", apiErr.ErrorCode(), "message", apiErr.ErrorMessage(), "error fault", apiErr.ErrorFault())
								_, err = j.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: j.queue, Attributes: j.attributes, Tags: j.tags})
								if err != nil {
									j.log.Error("create queue", "error", err)
								}
								// To successfully create a new queue, you must provide a
								// queue name that adheres to the limits related to queues
								// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/limits-queues.html)
								// and is unique within the scope of your queues. After you create a queue, you
								// must wait at least one second after the queue is created to be able to use the <------------
								// queue. To get the queue URL, use the GetQueueUrl action. GetQueueUrl require
								time.Sleep(time.Second * 2)
								continue
							}
						}
					}
				}

				j.log.Error("receive message", "error", err)
				continue
			}

			for i := 0; i < len(message.Messages); i++ {
				m := message.Messages[i]
				item, err := j.unpack(&m)
				if err != nil {
					_, errD := j.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
						QueueUrl:      j.queueURL,
						ReceiptHandle: m.ReceiptHandle,
					})
					if errD != nil {
						j.log.Error("message unpack, failed to delete the message from the queue", "error", err)
					}

					j.log.Error("message unpack", "error", err)
					continue
				}

				j.pq.Insert(item)
			}
		}
	}
}

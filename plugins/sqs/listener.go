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

func (c *consumer) listen(ctx context.Context) { //nolint:gocognit
	for {
		select {
		case <-c.pauseCh:
			c.log.Warn("sqs listener stopped")
			return
		default:
			message, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              c.queueURL,
				MaxNumberOfMessages:   c.prefetch,
				AttributeNames:        []types.QueueAttributeName{types.QueueAttributeName(ApproximateReceiveCount)},
				MessageAttributeNames: []string{All},
				// The new value for the message's visibility timeout (in seconds). Values range: 0
				// to 43200. Maximum: 12 hours.
				VisibilityTimeout: c.visibilityTimeout,
				WaitTimeSeconds:   c.waitTime,
			})

			if err != nil {
				if oErr, ok := (err).(*smithy.OperationError); ok {
					if rErr, ok := oErr.Err.(*http.ResponseError); ok {
						if apiErr, ok := rErr.Err.(*smithy.GenericAPIError); ok {
							// in case of NonExistentQueue - recreate the queue
							if apiErr.Code == NonExistentQueue {
								c.log.Error("receive message", "error code", apiErr.ErrorCode(), "message", apiErr.ErrorMessage(), "error fault", apiErr.ErrorFault())
								_, err = c.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: c.queue, Attributes: c.attributes, Tags: c.tags})
								if err != nil {
									c.log.Error("create queue", "error", err)
								}
								// To successfully create a new queue, you must provide a
								// queue name that adheres to the limits related to the queues
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

				c.log.Error("receive message", "error", err)
				continue
			}

			for i := 0; i < len(message.Messages); i++ {
				m := message.Messages[i]
				item, err := c.unpack(&m)
				if err != nil {
					_, errD := c.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
						QueueUrl:      c.queueURL,
						ReceiptHandle: m.ReceiptHandle,
					})
					if errD != nil {
						c.log.Error("message unpack, failed to delete the message from the queue", "error", err)
					}

					c.log.Error("message unpack", "error", err)
					continue
				}

				c.pq.Insert(item)
			}
		}
	}
}

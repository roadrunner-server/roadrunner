package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spiral/jobs/v2"
	"strconv"
	"time"
)

var jobAttributes = []*string{
	aws.String("rr-job"),
	aws.String("rr-maxAttempts"),
	aws.String("rr-delay"),
	aws.String("rr-timeout"),
	aws.String("rr-retryDelay"),
}

// pack job metadata into headers
func pack(url *string, j *jobs.Job) *sqs.SendMessageInput {
	return &sqs.SendMessageInput{
		QueueUrl:     url,
		DelaySeconds: aws.Int64(int64(j.Options.Delay)),
		MessageBody:  aws.String(j.Payload),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"rr-job":         {DataType: aws.String("String"), StringValue: aws.String(j.Job)},
			"rr-maxAttempts": {DataType: aws.String("String"), StringValue: awsString(j.Options.Attempts)},
			"rr-delay":       {DataType: aws.String("String"), StringValue: awsDuration(j.Options.DelayDuration())},
			"rr-timeout":     {DataType: aws.String("String"), StringValue: awsDuration(j.Options.TimeoutDuration())},
			"rr-retryDelay":  {DataType: aws.String("Number"), StringValue: awsDuration(j.Options.RetryDuration())},
		},
	}
}

// unpack restores jobs.Options
func unpack(msg *sqs.Message) (id string, attempt int, j *jobs.Job, err error) {
	if _, ok := msg.Attributes["ApproximateReceiveCount"]; !ok {
		return "", 0, nil, fmt.Errorf("missing attribute `%s`", "ApproximateReceiveCount")
	}
	attempt, _ = strconv.Atoi(*msg.Attributes["ApproximateReceiveCount"])

	for _, attr := range jobAttributes {
		if _, ok := msg.MessageAttributes[*attr]; !ok {
			return "", 0, nil, fmt.Errorf("missing message attribute `%s` (mixed queue?)", *attr)
		}
	}

	j = &jobs.Job{
		Job:     *msg.MessageAttributes["rr-job"].StringValue,
		Payload: *msg.Body,
		Options: &jobs.Options{},
	}

	if delay, err := strconv.Atoi(*msg.MessageAttributes["rr-delay"].StringValue); err == nil {
		j.Options.Delay = delay
	}

	if maxAttempts, err := strconv.Atoi(*msg.MessageAttributes["rr-maxAttempts"].StringValue); err == nil {
		j.Options.Attempts = maxAttempts
	}

	if timeout, err := strconv.Atoi(*msg.MessageAttributes["rr-timeout"].StringValue); err == nil {
		j.Options.Timeout = timeout
	}

	if retryDelay, err := strconv.Atoi(*msg.MessageAttributes["rr-retryDelay"].StringValue); err == nil {
		j.Options.RetryDelay = retryDelay
	}

	return *msg.MessageId, attempt - 1, j, nil
}

func awsString(n int) *string {
	return aws.String(strconv.Itoa(n))
}

func awsDuration(d time.Duration) *string {
	return aws.String(strconv.Itoa(int(d.Seconds())))
}

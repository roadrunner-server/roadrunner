package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Unpack(t *testing.T) {
	msg := &sqs.Message{
		Body:              aws.String("body"),
		Attributes:        map[string]*string{},
		MessageAttributes: map[string]*sqs.MessageAttributeValue{},
	}

	_, _, _, err := unpack(msg)
	assert.Error(t, err)
}

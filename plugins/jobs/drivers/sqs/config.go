package sqs

import "github.com/aws/aws-sdk-go-v2/aws"

const (
	attributes string = "attributes"
	tags       string = "tags"
	queue      string = "queue"
	pref       string = "prefetch"
	visibility string = "visibility_timeout"
	waitTime   string = "wait_time"
)

type GlobalCfg struct {
	Key          string `mapstructure:"key"`
	Secret       string `mapstructure:"secret"`
	Region       string `mapstructure:"region"`
	SessionToken string `mapstructure:"session_token"`
	Endpoint     string `mapstructure:"endpoint"`
}

// Config is used to parse pipeline configuration
type Config struct {
	// The duration (in seconds) that the received messages are hidden from subsequent
	// retrieve requests after being retrieved by a ReceiveMessage request.
	VisibilityTimeout int32 `mapstructure:"visibility_timeout"`
	// The duration (in seconds) for which the call waits for a message to arrive
	// in the queue before returning. If a message is available, the call returns
	// sooner than WaitTimeSeconds. If no messages are available and the wait time
	// expires, the call returns successfully with an empty list of messages.
	WaitTimeSeconds int32 `mapstructure:"wait_time_seconds"`
	// Prefetch is the maximum number of messages to return. Amazon SQS never returns more messages
	// than this value (however, fewer messages might be returned). Valid values: 1 to
	// 10. Default: 1.
	Prefetch int32 `mapstructure:"prefetch"`
	// The name of the new queue. The following limits apply to this name:
	//
	// * A queue
	// name can have up to 80 characters.
	//
	// * Valid values: alphanumeric characters,
	// hyphens (-), and underscores (_).
	//
	// * A FIFO queue name must end with the .fifo
	// suffix.
	//
	// Queue URLs and names are case-sensitive.
	//
	// This member is required.
	Queue *string `mapstructure:"queue"`

	// A map of attributes with their corresponding values. The following lists the
	// names, descriptions, and values of the special request parameters that the
	// CreateQueue action uses.
	// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SetQueueAttributes.html
	Attributes map[string]string `mapstructure:"attributes"`

	// From amazon docs:
	// Add cost allocation tags to the specified Amazon SQS queue. For an overview, see
	// Tagging Your Amazon SQS Queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-queue-tags.html)
	// in the Amazon SQS Developer Guide. When you use queue tags, keep the following
	// guidelines in mind:
	//
	// * Adding more than 50 tags to a queue isn't recommended.
	//
	// *
	// Tags don't have any semantic meaning. Amazon SQS interprets tags as character
	// strings.
	//
	// * Tags are case-sensitive.
	//
	// * A new tag with a key identical to that
	// of an existing tag overwrites the existing tag.
	//
	// For a full list of tag
	// restrictions, see Quotas related to queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-limits.html#limits-queues)
	// in the Amazon SQS Developer Guide. To be able to tag a queue on creation, you
	// must have the sqs:CreateQueue and sqs:TagQueue permissions. Cross-account
	// permissions don't apply to this action. For more information, see Grant
	// cross-account permissions to a role and a user name
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-customer-managed-policy-examples.html#grant-cross-account-permissions-to-role-and-user-name)
	// in the Amazon SQS Developer Guide.
	Tags map[string]string `mapstructure:"tags"`
}

func (c *GlobalCfg) InitDefault() {
	if c.Endpoint == "" {
		c.Endpoint = "http://localhost:9324"
	}
}

func (c *Config) InitDefault() {
	if c.Queue == nil {
		c.Queue = aws.String("default")
	}

	if c.Prefetch == 0 || c.Prefetch > 10 {
		c.Prefetch = 10
	}

	if c.WaitTimeSeconds == 0 {
		c.WaitTimeSeconds = 5
	}

	if c.Attributes == nil {
		c.Attributes = make(map[string]string)
	}

	if c.Tags == nil {
		c.Tags = make(map[string]string)
	}
}

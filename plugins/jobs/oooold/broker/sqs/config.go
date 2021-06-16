package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spiral/roadrunner/service"
	"time"
)

// Config defines sqs broker configuration.
type Config struct {
	// Region defined SQS region, not required when endpoint is not empty.
	Region string

	// Region defined AWS API key, not required when endpoint is not empty.
	Key string

	// Region defined AWS API secret, not required when endpoint is not empty.
	Secret string

	// Endpoint can be used to re-define SQS endpoint to custom location. Only for local development.
	Endpoint string

	// Timeout to allocate the connection. Default 10 seconds.
	Timeout int
}

// Hydrate config values.
func (c *Config) Hydrate(cfg service.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Region == "" {
		return fmt.Errorf("SQS region is missing")
	}

	if c.Key == "" {
		return fmt.Errorf("SQS key is missing")
	}

	if c.Secret == "" {
		return fmt.Errorf("SQS secret is missing")
	}

	return nil
}

// TimeoutDuration returns number of seconds allowed to allocate the connection.
func (c *Config) TimeoutDuration() time.Duration {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 10
	}

	return time.Duration(timeout) * time.Second
}

// Session returns new AWS session.
func (c *Config) Session() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.Key, c.Secret, ""),
	})
}

// SQS returns new SQS instance or error.
func (c *Config) SQS() (*sqs.SQS, error) {
	sess, err := c.Session()
	if err != nil {
		return nil, err
	}

	if c.Endpoint == "" {
		return sqs.New(sess), nil
	}

	return sqs.New(sess, &aws.Config{Endpoint: aws.String(c.Endpoint)}), nil
}

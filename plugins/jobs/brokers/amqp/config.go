package amqp

import "time"

// Config defines sqs broker configuration.
type Config struct {
	// Addr of AMQP server (example: amqp://guest:guest@localhost:5672/).
	Addr string

	// Timeout to allocate the connection. Default 10 seconds.
	Timeout int
}

// TimeoutDuration returns number of seconds allowed to redial
func (c *Config) TimeoutDuration() time.Duration {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 10
	}

	return time.Duration(timeout) * time.Second
}

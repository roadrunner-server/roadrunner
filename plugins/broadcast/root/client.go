package broadcast

import "sync"

// Client subscribes to a given topic and consumes or publish messages to it.
type Client struct {
	upstream chan *Message
	broker   Broker
	mu       sync.Mutex
	topics   []string
	patterns []string
}

// Channel returns incoming messages channel.
func (c *Client) Channel() chan *Message {
	return c.upstream
}

// Publish message into associated topic or topics.
func (c *Client) Publish(msg ...*Message) error {
	return c.broker.Publish(msg...)
}

// Subscribe client to specific topics.
func (c *Client) Subscribe(topics ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newTopics := make([]string, 0)
	for _, topic := range topics {
		found := false
		for _, e := range c.topics {
			if e == topic {
				found = true
				break
			}
		}

		if !found {
			newTopics = append(newTopics, topic)
		}
	}

	if len(newTopics) == 0 {
		return nil
	}

	c.topics = append(c.topics, newTopics...)

	return c.broker.Subscribe(c.upstream, newTopics...)
}

// SubscribePattern subscribe client to the specific topic pattern.
func (c *Client) SubscribePattern(pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, g := range c.patterns {
		if g == pattern {
			return nil
		}
	}

	c.patterns = append(c.patterns, pattern)
	return c.broker.SubscribePattern(c.upstream, pattern)
}

// Unsubscribe client from specific topics
func (c *Client) Unsubscribe(topics ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dropTopics := make([]string, 0)
	for _, topic := range topics {
		for i, e := range c.topics {
			if e == topic {
				c.topics = append(c.topics[:i], c.topics[i+1:]...)
				dropTopics = append(dropTopics, topic)
			}
		}
	}

	if len(dropTopics) == 0 {
		return nil
	}

	return c.broker.Unsubscribe(c.upstream, dropTopics...)
}

// UnsubscribePattern client from topic pattern.
func (c *Client) UnsubscribePattern(pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.patterns {
		if c.patterns[i] == pattern {
			c.patterns = append(c.patterns[:i], c.patterns[i+1:]...)

			return c.broker.UnsubscribePattern(c.upstream, pattern)
		}
	}

	return nil
}

// Topics return all the topics client subscribed to.
func (c *Client) Topics() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.topics
}

// Patterns return all the patterns client subscribed to.
func (c *Client) Patterns() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.patterns
}

// Close the client and consumption.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.topics) != 0 {
		err = c.broker.Unsubscribe(c.upstream, c.topics...)
	}

	close(c.upstream)
	return err
}

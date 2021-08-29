package amqpjobs

import (
	"github.com/spiral/errors"
)

func (c *consumer) initRabbitMQ() error {
	const op = errors.Op("jobs_plugin_rmq_init")
	// Channel opens a unique, concurrent server channel to process the bulk of AMQP
	// messages.  Any error from methods on this receiver will render the receiver
	// invalid and a new Channel should be opened.
	channel, err := c.conn.Channel()
	if err != nil {
		return errors.E(op, err)
	}

	// declare an exchange (idempotent operation)
	err = channel.ExchangeDeclare(
		c.exchangeName,
		c.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.E(op, err)
	}

	// verify or declare a queue
	q, err := channel.QueueDeclare(
		c.queue,
		false,
		false,
		c.exclusive,
		false,
		nil,
	)
	if err != nil {
		return errors.E(op, err)
	}

	// bind queue to the exchange
	err = channel.QueueBind(
		q.Name,
		c.routingKey,
		c.exchangeName,
		false,
		nil,
	)
	if err != nil {
		return errors.E(op, err)
	}

	return channel.Close()
}

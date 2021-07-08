package amqp

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

func (j *JobsConsumer) initRabbitMQ() (<-chan amqp.Delivery, error) {
	// Channel opens a unique, concurrent server channel to process the bulk of AMQP
	// messages.  Any error from methods on this receiver will render the receiver
	// invalid and a new Channel should be opened.
	channel, err := j.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.Qos(j.prefetchCount, 0, false)
	if err != nil {
		return nil, err
	}

	// declare an exchange (idempotent operation)
	err = channel.ExchangeDeclare(
		j.exchangeName,
		j.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// verify or declare a queue
	q, err := channel.QueueDeclare(
		fmt.Sprintf("%s.%s", j.routingKey, uuid.NewString()),
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// bind queue to the exchange
	err = channel.QueueBind(
		q.Name,
		j.routingKey,
		j.exchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// start reading messages from the channel
	deliv, err := channel.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return deliv, nil
}

func (j *JobsConsumer) listener(deliv <-chan amqp.Delivery) {
	go func() {
		for {
			select {
			case msg, ok := <-deliv:
				if !ok {
					j.logger.Info("delivery channel closed, leaving the rabbit listener")
					return
				}

				// add task to the queue
				j.pq.Insert(From(msg))
			case <-j.stop:
				return
			}
		}
	}()
}

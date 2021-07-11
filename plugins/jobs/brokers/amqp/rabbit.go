package amqp

import (
	"github.com/spiral/errors"
	"github.com/streadway/amqp"
)

func (j *JobsConsumer) initRabbitMQ() error {
	const op = errors.Op("rabbit_initmq")
	// Channel opens a unique, concurrent server channel to process the bulk of AMQP
	// messages.  Any error from methods on this receiver will render the receiver
	// invalid and a new Channel should be opened.
	channel, err := j.conn.Channel()
	if err != nil {
		return errors.E(op, err)
	}

	err = channel.Qos(j.prefetchCount, 0, false)
	if err != nil {
		return errors.E(op, err)
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
		return errors.E(op, err)
	}

	// verify or declare a queue
	q, err := channel.QueueDeclare(
		j.queue,
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return errors.E(op, err)
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
		return errors.E(op, err)
	}

	return channel.Close()
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

				d, err := FromDelivery(msg)
				if err != nil {
					j.logger.Error("amqp delivery convert", "error", err)
					continue
				}
				// insert job into the main priority queue
				j.pq.Insert(d)
			case <-j.stop:
				return
			}
		}
	}()
}

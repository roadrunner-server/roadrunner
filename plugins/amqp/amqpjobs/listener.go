package amqpjobs

import amqp "github.com/rabbitmq/amqp091-go"

func (c *consumer) listener(deliv <-chan amqp.Delivery) {
	go func() {
		for { //nolint:gosimple
			select {
			case msg, ok := <-deliv:
				if !ok {
					c.log.Info("delivery channel closed, leaving the rabbit listener")
					return
				}

				d, err := c.fromDelivery(msg)
				if err != nil {
					c.log.Error("amqp delivery convert", "error", err)
					continue
				}
				// insert job into the main priority queue
				c.pq.Insert(d)
			}
		}
	}()
}

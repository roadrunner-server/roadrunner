package amqp

import "github.com/streadway/amqp"

func (j *JobsConsumer) listener(deliv <-chan amqp.Delivery) {
	go func() {
		for { //nolint:gosimple
			select {
			case msg, ok := <-deliv:
				if !ok {
					j.log.Info("delivery channel closed, leaving the rabbit listener")
					return
				}

				d, err := fromDelivery(msg)
				if err != nil {
					j.log.Error("amqp delivery convert", "error", err)
					continue
				}
				// insert job into the main priority queue
				j.pq.Insert(d)
			}
		}
	}()
}

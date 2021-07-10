package amqp

import (
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"github.com/streadway/amqp"
)

// redialer used to redial to the rabbitmq in case of the connection interrupts
func (j *JobsConsumer) redialer() { //nolint:gocognit
	go func() {
		const op = errors.Op("rabbitmq_redial")
		for err := range j.conn.NotifyClose(make(chan *amqp.Error)) {
			if err != nil {
				j.Lock()

				j.logger.Error("connection closed, reconnecting", "error", err)

				expb := backoff.NewExponentialBackOff()
				// set the retry timeout (minutes)
				expb.MaxElapsedTime = j.retryTimeout
				op := func() error {
					j.logger.Warn("rabbitmq reconnecting, caused by", "error", err)
					var dialErr error
					j.conn, dialErr = amqp.Dial(j.connStr)
					if dialErr != nil {
						return fmt.Errorf("fail to dial server endpoint: %v", dialErr)
					}

					j.logger.Info("rabbitmq dial succeed. trying to redeclare queues and subscribers")

					// re-init connection
					errInit := j.initRabbitMQ()
					if errInit != nil {
						j.logger.Error("error while redialing", "error", errInit)
						return errInit
					}

					// redeclare consume channel
					var errConnCh error
					j.consumeChan, errConnCh = j.conn.Channel()
					if errConnCh != nil {
						return errors.E(op, errConnCh)
					}

					// redeclare publish channel
					var errPubCh error
					j.publishChan, errPubCh = j.conn.Channel()
					if errPubCh != nil {
						return errors.E(op, errPubCh)
					}

					// start reading messages from the channel
					deliv, err := j.consumeChan.Consume(
						j.queue,
						j.consumeID,
						false,
						false,
						false,
						false,
						nil,
					)
					if err != nil {
						return errors.E(op, err)
					}

					// restart listener
					j.listener(deliv)

					j.logger.Info("queues and subscribers redeclare succeed")
					return nil
				}

				retryErr := backoff.Retry(op, expb)
				if retryErr != nil {
					j.Unlock()
					j.logger.Error("backoff failed", "error", retryErr)
					return
				}

				j.Unlock()
			}
		}
	}()
}
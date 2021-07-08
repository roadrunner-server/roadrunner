package amqp

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

// redialer used to redial to the rabbitmq in case of the connection interrupts
func (j *JobsConsumer) redialer() {
	go func() {
		for err := range j.conn.NotifyClose(make(chan *amqp.Error)) {
			if err != nil {
				j.logger.Error("connection closed, reconnecting", "error", err)

				expb := backoff.NewExponentialBackOff()
				// set the retry timeout (minutes)
				expb.MaxElapsedTime = time.Minute * j.retryTimeout
				op := func() error {
					j.logger.Warn("rabbitmq reconnecting, caused by", "error", err)

					j.Lock()
					var dialErr error
					j.conn, dialErr = amqp.Dial(j.connStr)
					if dialErr != nil {
						j.Unlock()
						return fmt.Errorf("fail to dial server endpoint: %v", dialErr)
					}
					j.Unlock()

					j.logger.Info("rabbitmq dial succeed. trying to redeclare queues and subscribers")

					// re-init connection
					deliv, errInit := j.initRabbitMQ()
					if errInit != nil {
						j.Unlock()
						j.logger.Error("error while redialing", "error", errInit)
						return errInit
					}

					// restart listener
					j.listener(deliv)

					j.logger.Info("queues and subscribers redeclare succeed")
					return nil
				}

				retryErr := backoff.Retry(op, expb)
				if retryErr != nil {
					j.logger.Error("backoff failed", "error", retryErr)
					return
				}
			}
		}
	}()
}

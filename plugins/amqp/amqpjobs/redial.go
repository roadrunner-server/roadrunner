package amqpjobs

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

// redialer used to redial to the rabbitmq in case of the connection interrupts
func (c *consumer) redialer() { //nolint:gocognit
	go func() {
		const op = errors.Op("rabbitmq_redial")

		for {
			select {
			case err := <-c.conn.NotifyClose(make(chan *amqp.Error)):
				if err == nil {
					return
				}

				c.Lock()

				// trash the broken publishing channel
				<-c.publishChan

				t := time.Now().UTC()
				pipe := c.pipeline.Load().(*pipeline.Pipeline)

				c.eh.Push(events.JobEvent{
					Event:    events.EventPipeError,
					Pipeline: pipe.Name(),
					Driver:   pipe.Driver(),
					Error:    err,
					Start:    time.Now().UTC(),
				})

				expb := backoff.NewExponentialBackOff()
				// set the retry timeout (minutes)
				expb.MaxElapsedTime = c.retryTimeout
				operation := func() error {
					c.log.Warn("rabbitmq reconnecting, caused by", "error", err)
					var dialErr error
					c.conn, dialErr = amqp.Dial(c.connStr)
					if dialErr != nil {
						return errors.E(op, dialErr)
					}

					c.log.Info("rabbitmq dial succeed. trying to redeclare queues and subscribers")

					// re-init connection
					errInit := c.initRabbitMQ()
					if errInit != nil {
						c.log.Error("rabbitmq dial", "error", errInit)
						return errInit
					}

					// redeclare consume channel
					var errConnCh error
					c.consumeChan, errConnCh = c.conn.Channel()
					if errConnCh != nil {
						return errors.E(op, errConnCh)
					}

					// redeclare publish channel
					pch, errPubCh := c.conn.Channel()
					if errPubCh != nil {
						return errors.E(op, errPubCh)
					}

					// start reading messages from the channel
					deliv, err := c.consumeChan.Consume(
						c.queue,
						c.consumeID,
						false,
						false,
						false,
						false,
						nil,
					)
					if err != nil {
						return errors.E(op, err)
					}

					// put the fresh publishing channel
					c.publishChan <- pch
					// restart listener
					c.listener(deliv)

					c.log.Info("queues and subscribers redeclared successfully")

					return nil
				}

				retryErr := backoff.Retry(operation, expb)
				if retryErr != nil {
					c.Unlock()
					c.log.Error("backoff failed", "error", retryErr)
					return
				}

				c.eh.Push(events.JobEvent{
					Event:    events.EventPipeActive,
					Pipeline: pipe.Name(),
					Driver:   pipe.Driver(),
					Start:    t,
					Elapsed:  time.Since(t),
				})

				c.Unlock()

			case <-c.stopCh:
				if c.publishChan != nil {
					pch := <-c.publishChan
					err := pch.Close()
					if err != nil {
						c.log.Error("publish channel close", "error", err)
					}
				}

				if c.consumeChan != nil {
					err := c.consumeChan.Close()
					if err != nil {
						c.log.Error("consume channel close", "error", err)
					}
				}
				if c.conn != nil {
					err := c.conn.Close()
					if err != nil {
						c.log.Error("amqp connection close", "error", err)
					}
				}

				return
			}
		}
	}()
}

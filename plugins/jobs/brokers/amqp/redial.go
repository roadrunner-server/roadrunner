package amqp

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/streadway/amqp"
)

// redialer used to redial to the rabbitmq in case of the connection interrupts
func (j *JobsConsumer) redialer() { //nolint:gocognit
	go func() {
		const op = errors.Op("rabbitmq_redial")

		for {
			select {
			case err := <-j.conn.NotifyClose(make(chan *amqp.Error)):
				if err == nil {
					return
				}

				j.Lock()

				t := time.Now()
				pipe := j.pipeline.Load().(*pipeline.Pipeline)
				j.eh.Push(events.JobEvent{
					Event:    events.EventPipeError,
					Pipeline: pipe.Name(),
					Driver:   pipe.Driver(),
					Error:    err,
					Start:    time.Now(),
					Elapsed:  0,
				})

				j.log.Error("connection closed, reconnecting", "error", err)
				expb := backoff.NewExponentialBackOff()
				// set the retry timeout (minutes)
				expb.MaxElapsedTime = j.retryTimeout
				op := func() error {
					j.log.Warn("rabbitmq reconnecting, caused by", "error", err)
					var dialErr error
					j.conn, dialErr = amqp.Dial(j.connStr)
					if dialErr != nil {
						return fmt.Errorf("fail to dial server endpoint: %v", dialErr)
					}

					j.log.Info("rabbitmq dial succeed. trying to redeclare queues and subscribers")

					// re-init connection
					errInit := j.initRabbitMQ()
					if errInit != nil {
						j.log.Error("rabbitmq dial", "error", errInit)
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

					j.log.Info("queues and subscribers redeclared successfully")
					return nil
				}

				retryErr := backoff.Retry(op, expb)
				if retryErr != nil {
					j.Unlock()
					j.log.Error("backoff failed", "error", retryErr)
					return
				}

				j.eh.Push(events.JobEvent{
					Event:    events.EventPipeActive,
					Pipeline: pipe.Name(),
					Driver:   pipe.Driver(),
					Start:    t,
					Elapsed:  time.Since(t),
				})

				j.Unlock()

			case <-j.stopCh:
				err := j.publishChan.Close()
				if err != nil {
					j.log.Error("publish channel close", "error", err)
				}
				err = j.consumeChan.Close()
				if err != nil {
					j.log.Error("consume channel close", "error", err)
				}
				err = j.conn.Close()
				if err != nil {
					j.log.Error("amqp connection close", "error", err)
				}
			}
		}
	}()
}

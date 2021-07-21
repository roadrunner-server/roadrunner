package beanstalk

import (
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
)

func (j *JobConsumer) listen() { //nolint:gocognit
	const op = errors.Op("beanstalk_listen")
	for {
		select {
		case <-j.stopCh:
			j.log.Warn("beanstalk listener stopped")
			return
		default:
			id, body, err := j.pool.Reserve(j.reserveTimeout)
			if err != nil {
				// reserve timeout
				if connErr, ok := err.(beanstalk.ConnError); ok {
					switch connErr.Err {
					case beanstalk.ErrTimeout:
						j.log.Warn("timeout expired", "warn", connErr.Error())
						continue
					default:
						j.log.Error("beanstalk connection error", "error", connErr.Error())

						// backoff here
						expb := backoff.NewExponentialBackOff()
						// set the retry timeout (minutes)
						expb.MaxElapsedTime = time.Minute * 5

						operation := func() error {
							errR := j.pool.Redial()
							if errR != nil {
								return errors.E(op, errR)
							}

							j.log.Info("beanstalk redial was successful")
							// reassign a pool
							return nil
						}

						retryErr := backoff.Retry(operation, expb)
						if retryErr != nil {
							j.log.Error("beanstalk backoff failed, exiting from listener", "error", connErr, "retry error", retryErr)
							return
						}
						continue
					}
				}
				j.log.Error("beanstalk reserve", "error", err)
				continue
			}

			item := &Item{}
			err = unpack(id, body, j.pool.conn, item)
			if err != nil {
				j.log.Error("beanstalk unpack item", "error", err)
				continue
			}

			// insert job into the priority queue
			j.pq.Insert(item)
		}
	}
}

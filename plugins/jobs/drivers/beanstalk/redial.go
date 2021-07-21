package beanstalk

import (
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func (j *JobConsumer) redial() {
	for range j.reconnectCh {
		// backoff here
		expb := backoff.NewExponentialBackOff()
		// set the retry timeout (minutes)
		expb.MaxElapsedTime = time.Minute * 5

		op := func() error {
			err := j.pool.Redial()
			if err != nil {
				return err
			}

			j.log.Info("beanstalk redial was successful")
			// reassign a pool
			return nil
		}

		retryErr := backoff.Retry(op, expb)
		if retryErr != nil {
			j.log.Error("beanstalk backoff failed", "error", retryErr)
			continue
		}

		// restart listener
		if atomic.LoadUint32(&j.listeners) == 1 {
			// stop previous listener
			j.stopCh <- struct{}{}
			go j.listen()
		}
	}
}

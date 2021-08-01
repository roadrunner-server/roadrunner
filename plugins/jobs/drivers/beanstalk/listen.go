package beanstalk

import (
	"github.com/beanstalkd/go-beanstalk"
)

func (j *JobConsumer) listen() {
	for {
		select {
		case <-j.stopCh:
			j.log.Warn("beanstalk listener stopped")
			return
		default:
			id, body, err := j.pool.Reserve(j.reserveTimeout)
			if err != nil {
				if errB, ok := err.(beanstalk.ConnError); ok {
					switch errB.Err { //nolint:gocritic
					case beanstalk.ErrTimeout:
						j.log.Info("beanstalk reserve timeout", "warn", errB.Op)
						continue
					}
				}
				// in case of other error - continue
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

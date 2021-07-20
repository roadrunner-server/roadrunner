package beanstalk

import (
	"sync/atomic"

	"github.com/beanstalkd/go-beanstalk"
)

func (j *JobConsumer) redial() {
	for range j.reconnectCh {
		// backoff here

		j.Lock()

		var err error
		j.conn, err = beanstalk.DialTimeout(j.network, j.addr, j.tout)
		if err != nil {
			panic(err)
		}

		j.tube = beanstalk.NewTube(j.conn, j.tName)
		j.tubeSet = beanstalk.NewTubeSet(j.conn, j.tName)

		// restart listener
		if atomic.LoadUint32(&j.listeners) == 1 {
			go j.listen()
		}

		j.Unlock()
	}
}

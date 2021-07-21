package beanstalk

func (j *JobConsumer) listen() {
	for {
		select {
		case <-j.stopCh:
			j.log.Warn("beanstalk listener stopped")
			return
		default:
			id, body, err := j.pool.Reserve(j.reserveTimeout)
			if err != nil {
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

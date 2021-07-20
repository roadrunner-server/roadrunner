package beanstalk

func (j *JobConsumer) listen() {
	for {
		select {
		case <-j.stopCh:
			j.log.Warn("beanstalk listener stopped")
			return

		default:
			// lock used here to prevent consume from the broken connection
			j.Lock()

			id, body, err := j.tubeSet.Reserve(j.reserveTimeout)
			if err != nil {
				j.log.Error("beanstalk reserve", "error", err)
				j.Unlock()
				continue
			}

			item := &Item{}
			err = unpack(id, body, j.conn, item)
			if err != nil {
				j.log.Error("beanstalk unpack item", "error", err)
				j.Unlock()
				continue
			}

			// insert job into the priority queue
			j.pq.Insert(item)

			j.Unlock()
		}
	}
}

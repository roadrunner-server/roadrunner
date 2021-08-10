package amqp

// requeueListener should handle items passed to requeue
func (j *JobConsumer) requeueListener() {
	go func() {
		for { //nolint:gosimple
			select {
			case item, ok := <-j.requeueCh:
				if !ok {
					j.log.Info("requeue channel closed")
					return
				}

				pch := <-j.publishChan

				headers, err := pack(item.ID(), item)
				if err != nil {
					j.publishChan <- pch
					j.log.Error("requeue pack", "error", err)
					continue
				}

				err = j.handleItem(item, headers, pch)
				if err != nil {
					j.publishChan <- pch
					j.log.Error("requeue handle item", "error", err)
					continue
				}

				j.publishChan <- pch
			}
		}
	}()
}

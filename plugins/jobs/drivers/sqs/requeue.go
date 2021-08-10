package sqs

import "context"

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

				// TODO(rustatian): what context to use
				err := j.handleItem(context.TODO(), item)
				if err != nil {
					j.log.Error("requeue handle item", "error", err)
					continue
				}
			}
		}
	}()
}

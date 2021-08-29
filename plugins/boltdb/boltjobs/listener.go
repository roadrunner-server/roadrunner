package boltjobs

import (
	"fmt"
	"time"

	"github.com/spiral/roadrunner/v2/utils"
)

func (c *consumer) listener() {
	tt := time.NewTicker(time.Second)
	for {
		select {
		case <-c.stopCh:
			c.log.Warn("boltdb listener stopped")
			return
		case <-tt.C:
			tx, err := c.db.Begin(false)
			if err != nil {
				panic(err)
			}

			b := tx.Bucket(utils.AsBytes(PushBucket))

			cursor := b.Cursor()

			k, v := cursor.First()
			_ = k
			_ = v

			fmt.Println("foo")
		}
	}
}

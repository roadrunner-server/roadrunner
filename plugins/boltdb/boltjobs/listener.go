package boltjobs

import "time"

func (c *consumer) listener() {
	tt := time.NewTicker(time.Second)
	for {
		select {
		case <-tt.C:
			tx, err := c.db.Begin(false)
			if err != nil {
				panic(err)
			}
			//cursor := tx.Cursor()

			err = tx.Commit()
			if err != nil {
				panic(err)
			}
		}
	}
}

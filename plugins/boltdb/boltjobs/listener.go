package boltjobs

import (
	"bytes"
	"encoding/gob"
	"sync/atomic"
	"time"

	"github.com/spiral/roadrunner/v2/utils"
	bolt "go.etcd.io/bbolt"
)

func (c *consumer) listener() {
	tt := time.NewTicker(time.Millisecond * 10)
	defer tt.Stop()
	for {
		select {
		case <-c.stopCh:
			c.log.Warn("boltdb listener stopped")
			return
		case <-tt.C:
			if atomic.LoadUint64(c.active) >= uint64(c.prefetch) {
				time.Sleep(time.Second)
				continue
			}

			tx, err := c.db.Begin(true)
			if err != nil {
				c.log.Error("failed to begin writable transaction, job will be read on the next attempt", "error", err)
				continue
			}

			b := tx.Bucket(utils.AsBytes(PushBucket))
			inQb := tx.Bucket(utils.AsBytes(InQueueBucket))

			// get first item
			k, v := b.Cursor().First()
			if k == nil && v == nil {
				_ = tx.Commit()
				continue
			}

			buf := bytes.NewReader(v)
			dec := gob.NewDecoder(buf)

			item := &Item{}
			err = dec.Decode(item)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			err = inQb.Put(utils.AsBytes(item.ID()), v)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			// delete key from the PushBucket
			err = b.Delete(k)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			err = tx.Commit()
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			// attach pointer to the DB
			item.attachDB(c.db, c.active, c.delayed)
			// as the last step, after commit, put the item into the PQ
			c.pq.Insert(item)
		}
	}
}

func (c *consumer) delayedJobsListener() {
	tt := time.NewTicker(time.Millisecond * 100)
	defer tt.Stop()
	for {
		select {
		case <-c.stopCh:
			c.log.Warn("boltdb listener stopped")
			return
		case <-tt.C:
			tx, err := c.db.Begin(true)
			if err != nil {
				c.log.Error("failed to begin writable transaction, job will be read on the next attempt", "error", err)
				continue
			}

			delayB := tx.Bucket(utils.AsBytes(DelayBucket))
			inQb := tx.Bucket(utils.AsBytes(InQueueBucket))

			// get first item
			k, v := delayB.Cursor().First()
			if k == nil && v == nil {
				_ = tx.Commit()
				continue
			}

			t, err := time.Parse(time.RFC3339, utils.AsString(k))
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			if t.After(time.Now()) {
				_ = tx.Commit()
				continue
			}

			buf := bytes.NewReader(v)
			dec := gob.NewDecoder(buf)

			item := &Item{}
			err = dec.Decode(item)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			err = inQb.Put(utils.AsBytes(item.ID()), v)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			// delete key from the PushBucket
			err = delayB.Delete(k)
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			err = tx.Commit()
			if err != nil {
				c.rollback(err, tx)
				continue
			}

			// attach pointer to the DB
			item.attachDB(c.db, c.active, c.delayed)
			// as the last step, after commit, put the item into the PQ
			c.pq.Insert(item)
		}
	}
}

func (c *consumer) rollback(err error, tx *bolt.Tx) {
	errR := tx.Rollback()
	if errR != nil {
		c.log.Error("transaction commit error, rollback failed", "error", err, "rollback error", errR)
	}

	c.log.Error("transaction commit error, rollback succeed", "error", err)
}

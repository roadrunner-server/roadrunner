package boltjobs

import (
	"bytes"
	"encoding/gob"
	"sync/atomic"
	"time"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/jobs/job"
	"github.com/spiral/roadrunner/v2/utils"
	"go.etcd.io/bbolt"
)

type Item struct {
	// Job contains pluginName of job broker (usually PHP class).
	Job string `json:"job"`

	// Ident is unique identifier of the job, should be provided from outside
	Ident string `json:"id"`

	// Payload is string data (usually JSON) passed to Job broker.
	Payload string `json:"payload"`

	// Headers with key-values pairs
	Headers map[string][]string `json:"headers"`

	// Options contains set of PipelineOptions specific to job execution. Can be empty.
	Options *Options `json:"options,omitempty"`
}

// Options carry information about how to handle given job.
type Options struct {
	// Priority is job priority, default - 10
	// pointer to distinguish 0 as a priority and nil as priority not set
	Priority int64 `json:"priority"`

	// Pipeline manually specified pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// Delay defines time duration to delay execution for. Defaults to none.
	Delay int64 `json:"delay,omitempty"`

	// private
	db      *bbolt.DB
	active  *uint64
	delayed *uint64
}

func (i *Item) ID() string {
	return i.Ident
}

func (i *Item) Priority() int64 {
	return i.Options.Priority
}

func (i *Item) Body() []byte {
	return utils.AsBytes(i.Payload)
}

func (i *Item) Context() ([]byte, error) {
	ctx, err := json.Marshal(
		struct {
			ID       string              `json:"id"`
			Job      string              `json:"job"`
			Headers  map[string][]string `json:"headers"`
			Pipeline string              `json:"pipeline"`
		}{ID: i.Ident, Job: i.Job, Headers: i.Headers, Pipeline: i.Options.Pipeline},
	)

	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func (i *Item) Ack() error {
	const op = errors.Op("boltdb_item_ack")
	tx, err := i.Options.db.Begin(true)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err)
	}

	inQb := tx.Bucket(utils.AsBytes(InQueueBucket))
	err = inQb.Delete(utils.AsBytes(i.ID()))
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err)
	}

	if i.Options.Delay > 0 {
		atomic.AddUint64(i.Options.delayed, ^uint64(0))
	} else {
		atomic.AddUint64(i.Options.active, ^uint64(0))
	}

	return tx.Commit()
}

func (i *Item) Nack() error {
	const op = errors.Op("boltdb_item_ack")
	/*
		steps:
		1. begin tx
		2. get item by ID from the InQueueBucket (previously put in the listener)
		3. put it back to the PushBucket
		4. Delete it from the InQueueBucket
	*/
	tx, err := i.Options.db.Begin(true)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err)
	}

	inQb := tx.Bucket(utils.AsBytes(InQueueBucket))
	v := inQb.Get(utils.AsBytes(i.ID()))

	pushB := tx.Bucket(utils.AsBytes(PushBucket))

	err = pushB.Put(utils.AsBytes(i.ID()), v)
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err)
	}

	err = inQb.Delete(utils.AsBytes(i.ID()))
	if err != nil {
		_ = tx.Rollback()
		return errors.E(op, err)
	}

	return tx.Commit()
}

/*
Requeue algorithm:
1. Rewrite item headers and delay.
2. Begin writable transaction on attached to the item db.
3. Delete item from the InQueueBucket
4. Handle items with the delay:
   4.1. Get DelayBucket
   4.2. Make a key by adding the delay to the time.Now() in RFC3339 format
   4.3. Put this key with value to the DelayBucket
5. W/o delay, put the key with value to the PushBucket (requeue)
*/
func (i *Item) Requeue(headers map[string][]string, delay int64) error {
	const op = errors.Op("boltdb_item_requeue")
	i.Headers = headers
	i.Options.Delay = delay

	tx, err := i.Options.db.Begin(true)
	if err != nil {
		return errors.E(op, err)
	}

	inQb := tx.Bucket(utils.AsBytes(InQueueBucket))
	err = inQb.Delete(utils.AsBytes(i.ID()))
	if err != nil {
		return errors.E(op, i.rollback(err, tx))
	}

	// encode the item
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(i)
	val := make([]byte, buf.Len())
	copy(val, buf.Bytes())
	buf.Reset()

	if delay > 0 {
		delayB := tx.Bucket(utils.AsBytes(DelayBucket))
		tKey := time.Now().UTC().Add(time.Second * time.Duration(delay)).Format(time.RFC3339)

		if err != nil {
			return errors.E(op, i.rollback(err, tx))
		}

		err = delayB.Put(utils.AsBytes(tKey), val)
		if err != nil {
			return errors.E(op, i.rollback(err, tx))
		}

		return tx.Commit()
	}

	pushB := tx.Bucket(utils.AsBytes(PushBucket))
	if err != nil {
		return errors.E(op, i.rollback(err, tx))
	}

	err = pushB.Put(utils.AsBytes(i.ID()), val)
	if err != nil {
		return errors.E(op, i.rollback(err, tx))
	}

	return tx.Commit()
}

func (i *Item) attachDB(db *bbolt.DB, active, delayed *uint64) {
	i.Options.db = db
	i.Options.active = active
	i.Options.delayed = delayed
}

func (i *Item) rollback(err error, tx *bbolt.Tx) error {
	errR := tx.Rollback()
	if errR != nil {
		return errors.Errorf("transaction commit error: %v, rollback failed: %v", err, errR)
	}
	return errors.Errorf("transaction commit error: %v", err)
}

func fromJob(job *job.Job) *Item {
	return &Item{
		Job:     job.Job,
		Ident:   job.Ident,
		Payload: job.Payload,
		Headers: job.Headers,
		Options: &Options{
			Priority: job.Options.Priority,
			Pipeline: job.Options.Pipeline,
			Delay:    job.Options.Delay,
		},
	}
}

package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spiral/jobs/v2"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type queue struct {
	active       int32
	pipe         *jobs.Pipeline
	url          *string
	reserve      time.Duration
	lockReserved time.Duration

	// queue events
	lsn func(event int, ctx interface{})

	// stop channel
	wait chan interface{}

	// active operations
	muw sync.RWMutex
	wg  sync.WaitGroup

	// exec handlers
	execPool   chan jobs.Handler
	errHandler jobs.ErrorHandler
}

func newQueue(pipe *jobs.Pipeline, lsn func(event int, ctx interface{})) (*queue, error) {
	if pipe.String("queue", "") == "" {
		return nil, fmt.Errorf("missing `queue` parameter on sqs pipeline `%s`", pipe.Name())
	}

	return &queue{
		pipe:         pipe,
		reserve:      pipe.Duration("reserve", time.Second),
		lockReserved: pipe.Duration("lockReserved", 300*time.Second),
		lsn:          lsn,
	}, nil
}

// declareQueue declared queue
func (q *queue) declareQueue(s *sqs.SQS) (*string, error) {
	attr := make(map[string]*string)
	for k, v := range q.pipe.Map("declare") {
		if vs, ok := v.(string); ok {
			attr[k] = aws.String(vs)
		}

		if vi, ok := v.(int); ok {
			attr[k] = aws.String(strconv.Itoa(vi))
		}

		if vb, ok := v.(bool); ok {
			if vb {
				attr[k] = aws.String("true")
			} else {
				attr[k] = aws.String("false")
			}
		}
	}

	if len(attr) != 0 {
		r, err := s.CreateQueue(&sqs.CreateQueueInput{
			QueueName:  aws.String(q.pipe.String("queue", "")),
			Attributes: attr,
		})

		return r.QueueUrl, err
	}

	// no need to create (get existed)
	r, err := s.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(q.pipe.String("queue", ""))})
	if err != nil {
		return nil, err
	}

	return r.QueueUrl, nil
}

// serve consumers
func (q *queue) serve(s *sqs.SQS, tout time.Duration) {
	q.wait = make(chan interface{})
	atomic.StoreInt32(&q.active, 1)

	var errored bool
	for {
		messages, stop, err := q.consume(s)
		if err != nil {
			if errored {
				// reoccurring error
				time.Sleep(tout)
			} else {
				errored = true
				q.report(err)
			}

			continue
		}
		errored = false

		if stop {
			return
		}

		for _, msg := range messages {
			h := <-q.execPool
			go func(h jobs.Handler, msg *sqs.Message) {
				err := q.do(s, h, msg)
				q.execPool <- h
				q.wg.Done()
				q.report(err)
			}(h, msg)
		}
	}
}

// consume and allocate connection.
func (q *queue) consume(s *sqs.SQS) ([]*sqs.Message, bool, error) {
	q.muw.Lock()
	defer q.muw.Unlock()

	select {
	case <-q.wait:
		return nil, true, nil
	default:
		r, err := s.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:              q.url,
			MaxNumberOfMessages:   aws.Int64(int64(q.pipe.Integer("prefetch", 1))),
			WaitTimeSeconds:       aws.Int64(int64(q.reserve.Seconds())),
			VisibilityTimeout:     aws.Int64(int64(q.lockReserved.Seconds())),
			AttributeNames:        []*string{aws.String("ApproximateReceiveCount")},
			MessageAttributeNames: jobAttributes,
		})
		if err != nil {
			return nil, false, err
		}

		q.wg.Add(len(r.Messages))

		return r.Messages, false, nil
	}
}

// do single message
func (q *queue) do(s *sqs.SQS, h jobs.Handler, msg *sqs.Message) (err error) {
	id, attempt, j, err := unpack(msg)
	if err != nil {
		go q.deleteMessage(s, msg, err)
		return err
	}

	// block the job based on known timeout
	_, err = s.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          q.url,
		ReceiptHandle:     msg.ReceiptHandle,
		VisibilityTimeout: aws.Int64(int64(j.Options.TimeoutDuration().Seconds())),
	})
	if err != nil {
		go q.deleteMessage(s, msg, err)
		return err
	}

	err = h(id, j)
	if err == nil {
		return q.deleteMessage(s, msg, nil)
	}

	q.errHandler(id, j, err)

	if !j.Options.CanRetry(attempt) {
		return q.deleteMessage(s, msg, err)
	}

	// retry after specified duration
	_, err = s.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		QueueUrl:          q.url,
		ReceiptHandle:     msg.ReceiptHandle,
		VisibilityTimeout: aws.Int64(int64(j.Options.RetryDelay)),
	})

	return err
}

func (q *queue) deleteMessage(s *sqs.SQS, msg *sqs.Message, err error) error {
	_, drr := s.DeleteMessage(&sqs.DeleteMessageInput{QueueUrl: q.url, ReceiptHandle: msg.ReceiptHandle})
	return drr
}

// stop the queue consuming
func (q *queue) stop() {
	if atomic.LoadInt32(&q.active) == 0 {
		return
	}

	atomic.StoreInt32(&q.active, 0)

	close(q.wait)
	q.muw.Lock()
	q.wg.Wait()
	q.muw.Unlock()
}

// add job to the queue
func (q *queue) send(s *sqs.SQS, j *jobs.Job) (string, error) {
	r, err := s.SendMessage(pack(q.url, j))
	if err != nil {
		return "", err
	}

	return *r.MessageId, nil
}

// return queue stats
func (q *queue) stat(s *sqs.SQS) (stat *jobs.Stat, err error) {
	r, err := s.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: q.url,
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
			aws.String("ApproximateNumberOfMessagesDelayed"),
			aws.String("ApproximateNumberOfMessagesNotVisible"),
		},
	})

	if err != nil {
		return nil, err
	}

	stat = &jobs.Stat{InternalName: q.pipe.String("queue", "")}

	for a, v := range r.Attributes {
		if a == "ApproximateNumberOfMessages" {
			if v, err := strconv.Atoi(*v); err == nil {
				stat.Queue = int64(v)
			}
		}

		if a == "ApproximateNumberOfMessagesNotVisible" {
			if v, err := strconv.Atoi(*v); err == nil {
				stat.Active = int64(v)
			}
		}

		if a == "ApproximateNumberOfMessagesDelayed" {
			if v, err := strconv.Atoi(*v); err == nil {
				stat.Delayed = int64(v)
			}
		}
	}

	return stat, nil
}

// throw handles service, server and pool events.
func (q *queue) report(err error) {
	if err != nil {
		q.lsn(jobs.EventPipeError, &jobs.PipelineError{Pipeline: q.pipe, Caused: err})
	}
}

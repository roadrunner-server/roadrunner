package queue

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/spiral/roadrunner/v2/pkg/worker"
)

const (
	initialSize          = 1
	maxInitialSize       = 8
	maxInternalSliceSize = 10
)

type Node struct {
	w []worker.BaseProcess
	// LL
	n *Node
}

type Queue struct {
	mu sync.Mutex

	head *Node
	tail *Node

	curr uint64
	len  uint64

	sliceSize uint64
}

func NewQueue() *Queue {
	q := &Queue{
		mu:        sync.Mutex{},
		head:      nil,
		tail:      nil,
		curr:      0,
		len:       0,
		sliceSize: 0,
	}

	return q
}

func (q *Queue) Push(w worker.BaseProcess) {
	q.mu.Lock()

	if q.head == nil {
		h := newNode(initialSize)
		q.head = h
		q.tail = h
		q.sliceSize = maxInitialSize
	} else if uint64(len(q.tail.w)) >= atomic.LoadUint64(&q.sliceSize) {
		n := newNode(maxInternalSliceSize)
		q.tail.n = n
		q.tail = n
		q.sliceSize = maxInternalSliceSize
	}

	q.tail.w = append(q.tail.w, w)

	atomic.AddUint64(&q.len, 1)

	q.mu.Unlock()
}

func (q *Queue) Pop(ctx context.Context) (worker.BaseProcess, error) {
	q.mu.Lock()

	if q.head == nil {
		return nil, nil
	}

	w := q.head.w[q.curr]
	q.head.w[q.curr] = nil
	atomic.AddUint64(&q.len, ^uint64(0))
	atomic.AddUint64(&q.curr, 1)

	if atomic.LoadUint64(&q.curr) >= uint64(len(q.head.w)) {
		n := q.head.n
		q.head.n = nil
		q.head = n
		q.curr = 0
	}

	q.mu.Unlock()

	return w, nil
}

func (q *Queue) Remove(_ int64) {}

func (q *Queue) Destroy() {}

func newNode(capacity int) *Node {
	return &Node{w: make([]worker.BaseProcess, 0, capacity)}
}

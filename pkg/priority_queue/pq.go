package priorityqueue

import (
	"sync"

	priorityqueue "github.com/spiral/roadrunner/v2/common/priority_queue"
)

type PQ struct {
	sync.RWMutex
	bh *BinHeap
}

func NewPriorityQueue() *PQ {
	return &PQ{
		bh: NewBinHeap(),
	}
}

func (p *PQ) GetMax() priorityqueue.Item {
	p.Lock()
	defer p.Unlock()
	return p.bh.GetMax()
}

func (p *PQ) Insert(item priorityqueue.Item) {
	p.Lock()
	p.bh.Insert(item)
	p.Unlock()
}

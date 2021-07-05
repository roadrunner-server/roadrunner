package priorityqueue

import "sync"

type PQ struct {
	sync.RWMutex
	bh *BinHeap
}

func NewPriorityQueue() *PQ {
	return &PQ{
		bh: NewBinHeap(),
	}
}

func (p *PQ) Insert(item PQItem) {
	p.Lock()
	p.bh.Insert(item)
	p.Unlock()
}

func (p *PQ) Get() PQItem {
	p.Lock()
	defer p.Unlock()
	return p.bh.GetMax()
}

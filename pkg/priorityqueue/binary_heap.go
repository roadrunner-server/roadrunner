/*
binary heap (min-heap) algorithm used as a core for the priority queue
*/

package priorityqueue

import (
	"sync"
	"sync/atomic"
)

type BinHeap struct {
	items []Item
	// find a way to use pointer to the raw data
	len  uint64
	cond sync.Cond
}

func NewBinHeap() *BinHeap {
	return &BinHeap{
		items: make([]Item, 0, 100),
		len:   0,
		cond:  sync.Cond{L: &sync.Mutex{}},
	}
}

func (bh *BinHeap) fixUp() {
	k := len(bh.items) - 1
	p := (k - 1) >> 1 // k-1 / 2

	for k > 0 {
		cur, par := (bh.items)[k], (bh.items)[p]

		if *cur.Priority() < *par.Priority() {
			bh.swap(k, p)
			k = p
			p = (k - 1) >> 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) swap(i, j int) {
	(bh.items)[i], (bh.items)[j] = (bh.items)[j], (bh.items)[i]
}

func (bh *BinHeap) fixDown(curr, end int) {
	cOneIdx := (curr << 1) + 1
	for cOneIdx <= end {
		cTwoIdx := -1
		if (curr<<1)+2 <= end {
			cTwoIdx = (curr << 1) + 2
		}

		idxToSwap := cOneIdx
		// oh my, so unsafe
		if cTwoIdx > -1 && *(bh.items)[cTwoIdx].Priority() < *(bh.items)[cOneIdx].Priority() {
			idxToSwap = cTwoIdx
		}
		if *(bh.items)[idxToSwap].Priority() < *(bh.items)[curr].Priority() {
			bh.swap(curr, idxToSwap)
			curr = idxToSwap
			cOneIdx = (curr << 1) + 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) Insert(item Item) {
	bh.cond.L.Lock()
	bh.items = append(bh.items, item)

	// add len to the slice
	atomic.AddUint64(&bh.len, 1)

	// fix binary heap up
	bh.fixUp()
	bh.cond.L.Unlock()

	// signal the goroutine on wait
	bh.cond.Signal()
}

func (bh *BinHeap) GetMax() Item {
	bh.cond.L.Lock()
	defer bh.cond.L.Unlock()

	for atomic.LoadUint64(&bh.len) == 0 {
		bh.cond.Wait()
	}

	bh.swap(0, int(bh.len-1))

	item := (bh.items)[int(bh.len)-1]
	bh.items = (bh).items[0 : int(bh.len)-1]
	bh.fixDown(0, int(bh.len-2))

	// reduce len
	atomic.AddUint64(&bh.len, ^uint64(0))
	return item
}

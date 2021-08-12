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
	len    uint64
	maxLen uint64
	cond   sync.Cond
}

func NewBinHeap(maxLen uint64) *BinHeap {
	return &BinHeap{
		items:  make([]Item, 0, 1000),
		len:    0,
		maxLen: maxLen,
		cond:   sync.Cond{L: &sync.Mutex{}},
	}
}

func (bh *BinHeap) fixUp() {
	k := bh.len - 1
	p := (k - 1) >> 1 // k-1 / 2

	for k > 0 {
		cur, par := (bh.items)[k], (bh.items)[p]

		if cur.Priority() < par.Priority() {
			bh.swap(k, p)
			k = p
			p = (k - 1) >> 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) swap(i, j uint64) {
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
		if cTwoIdx > -1 && (bh.items)[cTwoIdx].Priority() < (bh.items)[cOneIdx].Priority() {
			idxToSwap = cTwoIdx
		}
		if (bh.items)[idxToSwap].Priority() < (bh.items)[curr].Priority() {
			bh.swap(uint64(curr), uint64(idxToSwap))
			curr = idxToSwap
			cOneIdx = (curr << 1) + 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) Len() uint64 {
	return atomic.LoadUint64(&bh.len)
}

func (bh *BinHeap) Insert(item Item) {
	bh.cond.L.Lock()

	// check the binary heap len before insertion
	if bh.Len() > bh.maxLen {
		// unlock the mutex to proceed to get-max
		bh.cond.L.Unlock()

		// signal waiting goroutines
		for bh.Len() > 0 {
			// signal waiting goroutines
			bh.cond.Signal()
		}
		// lock mutex to proceed inserting into the empty slice
		bh.cond.L.Lock()
	}

	bh.items = append(bh.items, item)

	// add len to the slice
	atomic.AddUint64(&bh.len, 1)

	// fix binary heap up
	bh.fixUp()
	bh.cond.L.Unlock()

	// signal the goroutine on wait
	bh.cond.Signal()
}

func (bh *BinHeap) ExtractMin() Item {
	bh.cond.L.Lock()

	// if len == 0, wait for the signal
	for bh.Len() == 0 {
		bh.cond.Wait()
	}

	bh.swap(0, bh.len-1)

	item := (bh.items)[int(bh.len)-1]
	bh.items = (bh).items[0 : int(bh.len)-1]
	bh.fixDown(0, int(bh.len-2))

	// reduce len
	atomic.AddUint64(&bh.len, ^uint64(0))

	bh.cond.L.Unlock()
	return item
}

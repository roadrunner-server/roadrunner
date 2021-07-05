/*
binary heap (min-heap) algorithm used as a core for the priority queue
*/

package priorityqueue

type BinHeap []PQItem

func NewBinHeap() *BinHeap {
	return &BinHeap{}
}

func (bh *BinHeap) fixUp() {
	k := len(*bh) - 1
	p := (k - 1) >> 1 // k-1 / 2

	for k > 0 {
		cur, par := (*bh)[k], (*bh)[p]

		if cur.Priority() < par.Priority() {
			bh.swap(k, p)
			k = p
			p = (k - 1) >> 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) swap(i, j int) {
	(*bh)[i], (*bh)[j] = (*bh)[j], (*bh)[i]
}

func (bh *BinHeap) fixDown(curr, end int) {
	cOneIdx := curr*2 + 1
	for cOneIdx <= end {
		cTwoIdx := -1
		if curr*2+2 <= end {
			cTwoIdx = curr*2 + 2
		}

		idxToSwap := cOneIdx
		if cTwoIdx > -1 && (*bh)[cTwoIdx].Priority() < (*bh)[cOneIdx].Priority() {
			idxToSwap = cTwoIdx
		}
		if (*bh)[idxToSwap].Priority() < (*bh)[curr].Priority() {
			bh.swap(curr, idxToSwap)
			curr = idxToSwap
			cOneIdx = curr*2 + 1
		} else {
			return
		}
	}
}

func (bh *BinHeap) Insert(item PQItem) {
	*bh = append(*bh, item)
	bh.fixUp()
}

func (bh *BinHeap) GetMax() PQItem {
	l := len(*bh)
	if l == 0 {
		return nil
	}

	bh.swap(0, l-1)

	item := (*bh)[l-1]
	*bh = (*bh)[0 : l-1]
	bh.fixDown(0, l-2)
	return item
}

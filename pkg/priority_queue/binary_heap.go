/*
binary heap (min-heap) algorithm used as a core for the priority queue
*/

package priorityqueue

type BinHeap []PQItem

func NewBinHeap() *BinHeap {
	return &BinHeap{}
}

func (bh *BinHeap) Init(items []PQItem) {
	arraySize := len(items) - 1

	for i := arraySize/2 - 1; i >= 0; i-- {
		bh.shiftDown(items, i, arraySize)
	}

	for i := arraySize - 1; i >= 1; i-- {
		items[0], items[i] = items[i], items[0]
		bh.shiftDown(items, 0, i-1)
	}
}

func (bh *BinHeap) shiftDown(numbers []PQItem, k, n int) {
	// k << 1 is equal to k*2
	for k<<1 <= n {
		j := k << 1

		if j < n && numbers[j].Priority() < numbers[j+1].Priority() {
			j++
		}

		if !(numbers[k].Priority() < numbers[j].Priority()) {
			break
		}

		numbers[k], numbers[j] = numbers[j], numbers[k]
		k = j
	}
}
func (bh *BinHeap) fix() {}

func (bh *BinHeap) Push(_ PQItem) {}

func (bh *BinHeap) Pop() PQItem {
	bh.fix()
	// get min
	return nil
}

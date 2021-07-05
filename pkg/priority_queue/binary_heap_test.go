package priorityqueue

import (
	"testing"

	priorityqueue "github.com/spiral/roadrunner/v2/common/priority_queue"
	"github.com/stretchr/testify/require"
)

type Test int

func (t Test) Ack() {
}

func (t Test) Nack() {
}

func (t Test) Body() []byte {
	return nil
}

func (t Test) Context() []byte {
	return nil
}

func (t Test) ID() string {
	return ""
}

func (t Test) Priority() uint64 {
	return uint64(t)
}

func TestBinHeap_Init(t *testing.T) {
	a := []priorityqueue.Item{Test(2), Test(23), Test(33), Test(44), Test(1), Test(2), Test(2), Test(2), Test(4), Test(6), Test(99)}

	bh := NewBinHeap()

	for i := 0; i < len(a); i++ {
		bh.Insert(a[i])
	}

	expected := []priorityqueue.Item{Test(1), Test(2), Test(2), Test(2), Test(2), Test(4), Test(6), Test(23), Test(33), Test(44), Test(99)}

	res := make([]priorityqueue.Item, 0, 12)

	for i := 0; i < 11; i++ {
		item := bh.GetMax()
		res = append(res, item)
	}

	require.Equal(t, expected, res)
}

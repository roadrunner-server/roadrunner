package priorityqueue

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

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

func (t Test) Context() ([]byte, error) {
	return nil, nil
}

func (t Test) ID() string {
	return "none"
}

func (t Test) Priority() uint64 {
	return uint64(t)
}

func TestBinHeap_Init(t *testing.T) {
	a := []Item{Test(2), Test(23), Test(33), Test(44), Test(1), Test(2), Test(2), Test(2), Test(4), Test(6), Test(99)}

	bh := NewBinHeap(0)

	for i := 0; i < len(a); i++ {
		bh.Insert(a[i])
	}

	expected := []Item{Test(1), Test(2), Test(2), Test(2), Test(2), Test(4), Test(6), Test(23), Test(33), Test(44), Test(99)}

	res := make([]Item, 0, 12)

	for i := 0; i < 11; i++ {
		item := bh.GetMax()
		res = append(res, item)
	}

	require.Equal(t, expected, res)
}

func TestNewPriorityQueue(t *testing.T) {
	insertsPerSec := uint64(0)
	getPerSec := uint64(0)
	stopCh := make(chan struct{}, 1)
	pq := NewBinHeap(1000)

	go func() {
		tt3 := time.NewTicker(time.Millisecond * 10)
		for {
			select {
			case <-tt3.C:
				require.Less(t, pq.Len(), uint64(1002))
			case <-stopCh:
				return
			}
		}
	}()

	go func() {
		tt := time.NewTicker(time.Second)

		for {
			select {
			case <-tt.C:
				fmt.Println(fmt.Sprintf("Insert per second: %d", atomic.LoadUint64(&insertsPerSec)))
				atomic.StoreUint64(&insertsPerSec, 0)
				fmt.Println(fmt.Sprintf("GetMax per second: %d", atomic.LoadUint64(&getPerSec)))
				atomic.StoreUint64(&getPerSec, 0)
			case <-stopCh:
				tt.Stop()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				pq.GetMax()
				atomic.AddUint64(&getPerSec, 1)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				pq.Insert(Test(rand.Int())) //nolint:gosec
				atomic.AddUint64(&insertsPerSec, 1)
			}
		}
	}()

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	stopCh <- struct{}{}
	stopCh <- struct{}{}
	stopCh <- struct{}{}
}

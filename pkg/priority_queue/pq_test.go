package priorityqueue

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	insertsPerSec := uint64(0)
	getPerSec := uint64(0)
	stopCh := make(chan struct{}, 1)
	pq := NewPriorityQueue()

	go func() {
		tt := time.NewTicker(time.Second)

		for {
			select {
			case <-tt.C:
				fmt.Println(fmt.Sprintf("GetMax per second: %d", atomic.LoadUint64(&getPerSec)))
				fmt.Println(fmt.Sprintf("Insert per second: %d", atomic.LoadUint64(&insertsPerSec)))
				atomic.StoreUint64(&getPerSec, 0)
				atomic.StoreUint64(&insertsPerSec, 0)
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
				it := pq.Get()
				if it == nil {
					continue
				}
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
}

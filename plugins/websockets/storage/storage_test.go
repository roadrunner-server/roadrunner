package storage

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const predifined = "chat-1-2"

func TestNewBST(t *testing.T) {
	// create a new bst
	g := NewStorage()

	for i := 0; i < 100; i++ {
		g.InsertMany(uuid.NewString(), []string{"comments"})
	}

	for i := 0; i < 100; i++ {
		g.InsertMany(uuid.NewString(), []string{"comments2"})
	}

	for i := 0; i < 100; i++ {
		g.InsertMany(uuid.NewString(), []string{"comments3"})
	}

	res := make(map[string]struct{}, 100)
	assert.Len(t, res, 0)

	// should be 100
	g.GetByPtr([]string{"comments"}, res)
	assert.Len(t, res, 100)

	res = make(map[string]struct{}, 100)
	assert.Len(t, res, 0)

	// should be 100
	g.GetByPtr([]string{"comments2"}, res)
	assert.Len(t, res, 100)

	res = make(map[string]struct{}, 100)
	assert.Len(t, res, 0)

	// should be 100
	g.GetByPtr([]string{"comments3"}, res)
	assert.Len(t, res, 100)
}

func BenchmarkGraph(b *testing.B) {
	g := NewStorage()

	for i := 0; i < 1000; i++ {
		uid := uuid.New().String()
		g.InsertMany(uuid.NewString(), []string{uid})
	}

	g.Insert(uuid.NewString(), predifined)

	b.ResetTimer()
	b.ReportAllocs()

	res := make(map[string]struct{})

	for i := 0; i < b.N; i++ {
		g.GetByPtr([]string{predifined}, res)

		for i := range res {
			delete(res, i)
		}
	}
}

func BenchmarkBigSearch(b *testing.B) {
	g1 := NewStorage()

	predefinedSlice := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		predefinedSlice = append(predefinedSlice, uuid.NewString())
	}
	if predefinedSlice == nil {
		b.FailNow()
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	res := make(map[string]struct{}, 333)

	for i := 0; i < b.N; i++ {
		g1.GetByPtr(predefinedSlice, res)

		for i := range res {
			delete(res, i)
		}
	}
}

func BenchmarkBigSearchWithRemoves(b *testing.B) {
	g1 := NewStorage()

	predefinedSlice := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		predefinedSlice = append(predefinedSlice, uuid.NewString())
	}
	if predefinedSlice == nil {
		b.FailNow()
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	go func() {
		tt := time.NewTicker(time.Microsecond)

		res := make(map[string]struct{}, 1000)
		for {
			select {
			case <-tt.C:
				num := rand.Intn(1000) //nolint:gosec
				g1.GetByPtr(predefinedSlice, res)
				for k := range res {
					g1.Remove(k, predefinedSlice[num])
				}
			}
		}
	}()

	res := make(map[string]struct{}, 100)

	for i := 0; i < b.N; i++ {
		g1.GetByPtr(predefinedSlice, res)

		for i := range res {
			delete(res, i)
		}
	}
}

func TestBigSearchWithRemoves(t *testing.T) {
	g1 := NewStorage()

	predefinedSlice := make([]string, 0, 1000)
	for i := 0; i < 1000; i++ {
		predefinedSlice = append(predefinedSlice, uuid.NewString())
	}
	if predefinedSlice == nil {
		t.FailNow()
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 1000; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	stopCh := make(chan struct{})

	go func() {
		tt := time.NewTicker(time.Microsecond)

		res := make(map[string]struct{}, 1000)
		for {
			select {
			case <-tt.C:
				num := rand.Intn(1000) //nolint:gosec
				g1.GetByPtr(predefinedSlice, res)
				for k := range res {
					g1.Remove(k, predefinedSlice[num])
				}

			case <-stopCh:
				tt.Stop()
				return
			}
		}
	}()

	res := make(map[string]struct{}, 100)

	for i := 0; i < 1000; i++ {
		g1.GetByPtr(predefinedSlice, res)

		for i := range res {
			delete(res, i)
		}
	}

	stopCh <- struct{}{}
}

func TestGraph(t *testing.T) {
	g := NewStorage()

	for i := 0; i < 1000; i++ {
		uid := uuid.New().String()
		g.Insert(uuid.NewString(), uid)
	}

	g.Insert(uuid.NewString(), predifined)

	res := make(map[string]struct{})

	g.GetByPtr([]string{predifined}, res)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 1)
}

func TestTreeConcurrentContains(t *testing.T) {
	g := NewStorage()

	key1 := uuid.NewString()
	key2 := uuid.NewString()
	key3 := uuid.NewString()
	key4 := uuid.NewString()
	key5 := uuid.NewString()

	g.Insert(key1, predifined)
	g.Insert(key2, predifined)
	g.Insert(key3, predifined)
	g.Insert(key4, predifined)
	g.Insert(key5, predifined)

	res := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		go func() {
			g.GetByPtrTS([]string{predifined}, res)
		}()

		go func() {
			g.GetByPtrTS([]string{predifined}, res)
		}()

		go func() {
			g.GetByPtrTS([]string{predifined}, res)
		}()

		go func() {
			g.GetByPtrTS([]string{predifined}, res)
		}()
	}

	time.Sleep(time.Second * 5)

	res2 := make(map[string]struct{}, 5)

	g.GetByPtr([]string{predifined}, res2)
	assert.NotEmpty(t, res2)
	assert.Len(t, res2, 5)
}

func TestGraphRemove(t *testing.T) {
	g := NewStorage()

	key1 := uuid.NewString()
	key2 := uuid.NewString()
	key3 := uuid.NewString()
	key4 := uuid.NewString()
	key5 := uuid.NewString()

	g.Insert(key1, predifined)
	g.Insert(key2, predifined)
	g.Insert(key3, predifined)
	g.Insert(key4, predifined)
	g.Insert(key5, predifined)

	res := make(map[string]struct{}, 5)
	g.GetByPtr([]string{predifined}, res)
	assert.NotEmpty(t, res)
	assert.Len(t, res, 5)

	g.Remove(key1, predifined)

	res2 := make(map[string]struct{}, 4)
	g.GetByPtr([]string{predifined}, res2)
	assert.NotEmpty(t, res2)
	assert.Len(t, res2, 4)
}

package bst

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
	g := NewBST()

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments")
	}

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments2")
	}

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments3")
	}

	// should be 100
	exist := g.Get("comments")
	assert.Len(t, exist, 100)

	// should be 100
	exist2 := g.Get("comments2")
	assert.Len(t, exist2, 100)

	// should be 100
	exist3 := g.Get("comments3")
	assert.Len(t, exist3, 100)
}

func BenchmarkGraph(b *testing.B) {
	g := NewBST()

	for i := 0; i < 1000; i++ {
		uid := uuid.New().String()
		g.Insert(uuid.NewString(), uid)
	}

	g.Insert(uuid.NewString(), predifined)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		exist := g.Get(predifined)
		_ = exist
	}
}

func BenchmarkBigSearch(b *testing.B) {
	g1 := NewBST()
	g2 := NewBST()
	g3 := NewBST()

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
		g2.Insert(uuid.NewString(), uuid.NewString())
	}
	for i := 0; i < 1000; i++ {
		g3.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 333; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	for i := 0; i < 333; i++ {
		g2.Insert(uuid.NewString(), predefinedSlice[333+i])
	}

	for i := 0; i < 333; i++ {
		g3.Insert(uuid.NewString(), predefinedSlice[666+i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g1.Get(predefinedSlice[i])
			_ = exist
		}
	}
	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g2.Get(predefinedSlice[333+i])
			_ = exist
		}
	}
	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g3.Get(predefinedSlice[666+i])
			_ = exist
		}
	}
}

func BenchmarkBigSearchWithRemoves(b *testing.B) {
	g1 := NewBST()
	g2 := NewBST()
	g3 := NewBST()

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
		g2.Insert(uuid.NewString(), uuid.NewString())
	}
	for i := 0; i < 1000; i++ {
		g3.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 333; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	for i := 0; i < 333; i++ {
		g2.Insert(uuid.NewString(), predefinedSlice[333+i])
	}

	for i := 0; i < 333; i++ {
		g3.Insert(uuid.NewString(), predefinedSlice[666+i])
	}

	go func() {
		tt := time.NewTicker(time.Millisecond)
		for {
			select {
			case <-tt.C:
				num := rand.Intn(333) //nolint:gosec
				values := g1.Get(predefinedSlice[num])
				for k := range values {
					g1.Remove(k, predefinedSlice[num])
				}
			}
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g1.Get(predefinedSlice[i])
			_ = exist
		}
	}
	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g2.Get(predefinedSlice[333+i])
			_ = exist
		}
	}
	for i := 0; i < b.N; i++ {
		for i := 0; i < 333; i++ {
			exist := g3.Get(predefinedSlice[666+i])
			_ = exist
		}
	}
}

func TestGraph(t *testing.T) {
	g := NewBST()

	for i := 0; i < 1000; i++ {
		uid := uuid.New().String()
		g.Insert(uuid.NewString(), uid)
	}

	g.Insert(uuid.NewString(), predifined)

	exist := g.Get(predifined)
	assert.NotNil(t, exist)
	assert.Len(t, exist, 1)
}

func TestTreeConcurrentContains(t *testing.T) {
	g := NewBST()

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

	for i := 0; i < 100; i++ {
		go func() {
			_ = g.Get(predifined)
		}()

		go func() {
			_ = g.Get(predifined)
		}()

		go func() {
			_ = g.Get(predifined)
		}()

		go func() {
			_ = g.Get(predifined)
		}()
	}

	time.Sleep(time.Second * 2)

	exist := g.Get(predifined)
	assert.NotNil(t, exist)
	assert.Len(t, exist, 5)
}

func TestGraphRemove(t *testing.T) {
	g := NewBST()

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

	exist := g.Get(predifined)
	assert.NotNil(t, exist)
	assert.Len(t, exist, 5)

	g.Remove(key1, predifined)

	exist = g.Get(predifined)
	assert.NotNil(t, exist)
	assert.Len(t, exist, 4)
}

func TestBigSearch(t *testing.T) {
	g1 := NewBST()
	g2 := NewBST()
	g3 := NewBST()

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
		g2.Insert(uuid.NewString(), uuid.NewString())
	}
	for i := 0; i < 1000; i++ {
		g3.Insert(uuid.NewString(), uuid.NewString())
	}

	for i := 0; i < 333; i++ {
		g1.Insert(uuid.NewString(), predefinedSlice[i])
	}

	for i := 0; i < 333; i++ {
		g2.Insert(uuid.NewString(), predefinedSlice[333+i])
	}

	for i := 0; i < 333; i++ {
		g3.Insert(uuid.NewString(), predefinedSlice[666+i])
	}

	for i := 0; i < 333; i++ {
		exist := g1.Get(predefinedSlice[i])
		assert.NotNil(t, exist)
		assert.Len(t, exist, 1)
	}

	for i := 0; i < 333; i++ {
		exist := g2.Get(predefinedSlice[333+i])
		assert.NotNil(t, exist)
		assert.Len(t, exist, 1)
	}

	for i := 0; i < 333; i++ {
		exist := g3.Get(predefinedSlice[666+i])
		assert.NotNil(t, exist)
		assert.Len(t, exist, 1)
	}
}

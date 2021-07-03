package priorityqueue

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type Test int

func (t Test) ID() string {
	return ""
}

func (t Test) Priority() uint64 {
	return uint64(t)
}

func TestBinHeap_Init(t *testing.T) {
	a := []PQItem{Test(2), Test(23), Test(33), Test(44), Test(1), Test(2), Test(2), Test(2), Test(4), Test(6), Test(99)}

	bh := NewBinHeap()

	bh.Init(a)

	expected := []PQItem{Test(1), Test(2), Test(2), Test(2), Test(2), Test(4), Test(6), Test(23), Test(33), Test(44), Test(99)}

	require.Equal(t, expected, a)
}

func BenchmarkBinHeap_Init(b *testing.B) {
	a := []PQItem{Test(2), Test(23), Test(33), Test(44), Test(1), Test(2), Test(2), Test(2), Test(4), Test(6), Test(99)}
	bh := NewBinHeap()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bh.Init(a)
	}
}

func BenchmarkBinHeap_InitStdSort(b *testing.B) {
	a := []PQItem{Test(2), Test(23), Test(33), Test(44), Test(1), Test(2), Test(2), Test(2), Test(4), Test(6), Test(99)}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sort.Slice(a, func(i, j int) bool {
			return a[i].Priority() < a[j].Priority()
		})
	}
}

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFTime_UnixNano(t *testing.T) {
	ft := NewFastTime(time.Millisecond)
	defer ft.Stop()

	var d int64

	d = time.Now().UnixNano() - ft.UnixNano()

	assert.True(t, d >= 0)
	assert.True(t, d <= int64(time.Millisecond*2))

	time.Sleep(time.Millisecond * 100)
	d = time.Now().UnixNano() - ft.UnixNano()

	assert.True(t, d >= 0)
	assert.True(t, d <= int64(time.Millisecond*2))
}

func Benchmark_FastTime(b *testing.B) {
	ft := NewFastTime(time.Microsecond)
	defer ft.Stop()

	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = ft.UnixNano()
	}
}

func Benchmark_Time(b *testing.B) {
	ft := NewFastTime(time.Microsecond)
	defer ft.Stop()

	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = time.Now().UnixNano()
	}
}

package structs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions_CanRetry(t *testing.T) {
	opts := &Options{Attempts: 0}

	assert.False(t, opts.CanRetry(0))
	assert.False(t, opts.CanRetry(1))
}

func TestOptions_CanRetry_SameValue(t *testing.T) {
	opts := &Options{Attempts: 1}

	assert.False(t, opts.CanRetry(0))
	assert.False(t, opts.CanRetry(1))
}

func TestOptions_CanRetry_Value(t *testing.T) {
	opts := &Options{Attempts: 2}

	assert.True(t, opts.CanRetry(0))
	assert.False(t, opts.CanRetry(1))
	assert.False(t, opts.CanRetry(2))
}

func TestOptions_CanRetry_Value3(t *testing.T) {
	opts := &Options{Attempts: 3}

	assert.True(t, opts.CanRetry(0))
	assert.True(t, opts.CanRetry(1))
	assert.False(t, opts.CanRetry(2))
}

func TestOptions_RetryDuration(t *testing.T) {
	opts := &Options{RetryDelay: 0}
	assert.Equal(t, time.Duration(0), opts.RetryDuration())
}

func TestOptions_RetryDuration2(t *testing.T) {
	opts := &Options{RetryDelay: 1}
	assert.Equal(t, time.Second, opts.RetryDuration())
}

func TestOptions_DelayDuration(t *testing.T) {
	opts := &Options{Delay: 0}
	assert.Equal(t, time.Duration(0), opts.DelayDuration())
}

func TestOptions_DelayDuration2(t *testing.T) {
	opts := &Options{Delay: 1}
	assert.Equal(t, time.Second, opts.DelayDuration())
}

func TestOptions_TimeoutDuration(t *testing.T) {
	opts := &Options{Timeout: 0}
	assert.Equal(t, time.Minute*30, opts.TimeoutDuration())
}

func TestOptions_TimeoutDuration2(t *testing.T) {
	opts := &Options{Timeout: 1}
	assert.Equal(t, time.Second, opts.TimeoutDuration())
}

func TestOptions_Merge(t *testing.T) {
	opts := &Options{}

	opts.Merge(&Options{
		Pipeline:   "pipeline",
		Delay:      2,
		Timeout:    1,
		Attempts:   1,
		RetryDelay: 1,
	})

	assert.Equal(t, "pipeline", opts.Pipeline)
	assert.Equal(t, uint64(1), opts.Attempts)
	assert.Equal(t, uint64(2), opts.Delay)
	assert.Equal(t, uint64(1), opts.Timeout)
	assert.Equal(t, uint64(1), opts.RetryDelay)
}

func TestOptions_MergeKeepOriginal(t *testing.T) {
	opts := &Options{
		Pipeline:   "default",
		Delay:      10,
		Timeout:    10,
		Attempts:   10,
		RetryDelay: 10,
	}

	opts.Merge(&Options{
		Pipeline:   "pipeline",
		Delay:      2,
		Timeout:    1,
		Attempts:   1,
		RetryDelay: 1,
	})

	assert.Equal(t, "default", opts.Pipeline)
	assert.Equal(t, uint64(10), opts.Attempts)
	assert.Equal(t, uint64(10), opts.Delay)
	assert.Equal(t, uint64(10), opts.Timeout)
	assert.Equal(t, uint64(10), opts.RetryDelay)
}

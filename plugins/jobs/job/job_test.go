package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions_DelayDuration(t *testing.T) {
	opts := &Options{Delay: 0}
	assert.Equal(t, time.Duration(0), opts.DelayDuration())
}

func TestOptions_DelayDuration2(t *testing.T) {
	opts := &Options{Delay: 1}
	assert.Equal(t, time.Second, opts.DelayDuration())
}

func TestOptions_Merge(t *testing.T) {
	opts := &Options{}

	opts.Merge(&Options{
		Pipeline: "pipeline",
		Delay:    2,
	})

	assert.Equal(t, "pipeline", opts.Pipeline)
	assert.Equal(t, int64(2), opts.Delay)
}

func TestOptions_MergeKeepOriginal(t *testing.T) {
	opts := &Options{
		Pipeline: "default",
		Delay:    10,
	}

	opts.Merge(&Options{
		Pipeline: "pipeline",
		Delay:    2,
	})

	assert.Equal(t, "default", opts.Pipeline)
	assert.Equal(t, int64(10), opts.Delay)
}

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

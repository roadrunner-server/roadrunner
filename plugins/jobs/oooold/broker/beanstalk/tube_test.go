package beanstalk

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTube_CantServe(t *testing.T) {
	var gctx interface{}
	tube := &tube{
		lsn: func(event int, ctx interface{}) {
			gctx = ctx
		},
	}

	tube.serve(&Config{Addr: "broken"})
	assert.Error(t, gctx.(error))
}

package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sirupsen/logrus"
)

func TestContainer_Init(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	c.Register("test2", struct{}{})

	assert.Equal(t, 2, len(hook.Entries))

	assert.NoError(t, c.Serve())
	c.Stop()
}
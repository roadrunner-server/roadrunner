package log

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRpcServer_Log(t *testing.T) {
	c, hook := initContainer()
	err := c.Init(&testCfg{rpcCfg: `{"enable":true}`, httpCfg: `{}`})

	require.NoError(t, err)

	rpcSvc, status := c.Get(rpc.ID)
	assert.Equal(t, service.StatusOK, status)

	rpc, ok := rpcSvc.(*rpc.Service)

	assert.True(t, ok)

	go c.Serve()
	defer c.Stop()

	time.Sleep(time.Millisecond * 100)

	client, err := rpc.Client()

	assert.NoError(t, err)

	t.Run("Simple log", func(t *testing.T) {
		var response bool

		err := client.Call("log.Log", Entry{
			Level:   "warn",
			Message: "message",
		}, &response)

		assert.NoError(t, err)

		lastLog := hook.LastEntry()

		assert.True(t, response)
		if assert.NotNil(t, lastLog) {
			assert.Equal(t, "message", lastLog.Message)
			assert.Equal(t, logrus.WarnLevel, lastLog.Level)
			assert.Empty(t, lastLog.Data)
		}

	})

	t.Run("String as fields", func(t *testing.T) {
		var response bool

		err := client.Call("log.Log", Entry{
			Level:   "trace",
			Message: "message",
			Fields:  "foo",
		}, &response)

		assert.NoError(t, err)

		lastLog := hook.LastEntry()
		assert.True(t, response)

		if assert.NotNil(t, lastLog) {
			assert.Equal(t, "message", lastLog.Message)
			assert.Equal(t, logrus.TraceLevel, lastLog.Level)
			assert.Equal(t, "foo", lastLog.Data["data"])
		}
	})

	t.Run("Map as fields", func(t *testing.T) {
		var response bool

		err := client.Call("log.Log", Entry{
			Level:   "error",
			Message: "message",
			Fields:  map[string]interface{}{"foo": 1},
		}, &response)

		assert.NoError(t, err)

		lastLog := hook.LastEntry()
		assert.True(t, response)
		if assert.NotNil(t, lastLog) {
			assert.Equal(t, "message", lastLog.Message)
			assert.Equal(t, logrus.ErrorLevel, lastLog.Level)
			assert.Equal(t, float64(1), lastLog.Data["foo"])
		}
	})

	t.Run("Wrong level", func(t *testing.T) {
		hook.Reset()

		var response bool

		err := client.Call("log.Log", Entry{
			Level: "unknown",
		}, &response)

		assert.Error(t, err)
		assert.False(t, response)
		assert.Nil(t, hook.LastEntry())
	})
}

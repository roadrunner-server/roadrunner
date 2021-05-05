package old

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestRedis_Error(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	//c := service.NewContainer(logger)
	//c.Register(rpc.ID, &rpc.Service{})
	//c.Register(ID, &Service{})
	//
	//err := c.Init(&testCfg{
	//	broadcast: `{"redis":{"addr":"localhost:6372"}}`,
	//	rpc:       fmt.Sprintf(`{"join":"tcp://:%v"}`, rpcPort),
	//})

	rpcPort++

	assert.Error(t, err)
}

func TestRedis_Broadcast(t *testing.T) {
	br, _, c := setup(`{"redis":{"addr":"localhost:6379"}}`)
	defer c.Stop()

	client := br.NewClient()
	defer client.Close()

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello1"))) // must not be delivered

	assert.NoError(t, client.Subscribe("topic"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello1")))
	assert.Equal(t, `hello1`, readStr(<-client.Channel()))

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello2")))
	assert.Equal(t, `hello2`, readStr(<-client.Channel()))

	assert.NoError(t, client.Unsubscribe("topic"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello3")))

	assert.NoError(t, client.Subscribe("topic"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello4")))
	assert.Equal(t, `hello4`, readStr(<-client.Channel()))
}

func TestRedis_BroadcastPattern(t *testing.T) {
	br, _, c := setup(`{"redis":{"addr":"localhost:6379"}}`)
	defer c.Stop()

	client := br.NewClient()
	defer client.Close()

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello1"))) // must not be delivered

	assert.NoError(t, client.SubscribePattern("topic/*"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic/1", "hello1")))
	assert.Equal(t, `hello1`, readStr(<-client.Channel()))

	assert.NoError(t, br.Broker().Publish(newMessage("topic/2", "hello2")))
	assert.Equal(t, `hello2`, readStr(<-client.Channel()))

	assert.NoError(t, br.Broker().Publish(newMessage("different", "hello4")))
	assert.NoError(t, br.Broker().Publish(newMessage("topic/2", "hello5")))

	assert.Equal(t, `hello5`, readStr(<-client.Channel()))

	assert.NoError(t, client.UnsubscribePattern("topic/*"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic/3", "hello6")))

	assert.NoError(t, client.SubscribePattern("topic/*"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic/4", "hello7")))
	assert.Equal(t, `hello7`, readStr(<-client.Channel()))
}

func TestRedis_NotActive(t *testing.T) {
	b := &Redis{}
	b.stopped = 1

	assert.Error(t, b.Publish(nil))
	assert.Error(t, b.Subscribe(nil))
	assert.Error(t, b.Unsubscribe(nil))
	assert.Error(t, b.SubscribePattern(nil, ""))
	assert.Error(t, b.UnsubscribePattern(nil, ""))
}

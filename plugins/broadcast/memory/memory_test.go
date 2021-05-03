package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory_Broadcast(t *testing.T) {
	br, _, c := setup(`{}`)
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

func TestMemory_BroadcastPattern(t *testing.T) {
	br, _, c := setup(`{}`)
	defer c.Stop()

	client := br.NewClient()
	defer client.Close()

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello1"))) // must not be delivered

	assert.NoError(t, client.SubscribePattern("topic/*"))

	assert.NoError(t, br.Broker().Publish(newMessage("topic/1", "hello1")))
	assert.Equal(t, `hello1`, readStr(<-client.Channel()))

	assert.NoError(t, client.Publish(newMessage("topic/1", "hello1")))
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

func TestMemory_NotActive(t *testing.T) {
	b := memoryBroker()
	b.stopped = 1

	assert.Error(t, b.Publish(nil))
	assert.Error(t, b.Subscribe(nil))
	assert.Error(t, b.Unsubscribe(nil))
	assert.Error(t, b.SubscribePattern(nil, ""))
	assert.Error(t, b.UnsubscribePattern(nil, ""))
}

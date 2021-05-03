package broadcast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Client_Topics(t *testing.T) {
	br, _, c := setup(`{}`)
	defer c.Stop()

	client := br.NewClient()
	defer client.Close()

	assert.Equal(t, []string{}, client.Topics())

	assert.NoError(t, client.Subscribe("topic"))
	assert.Equal(t, []string{"topic"}, client.Topics())

	assert.NoError(t, client.Subscribe("topic"))
	assert.Equal(t, []string{"topic"}, client.Topics())

	assert.NoError(t, br.broker.Subscribe(client.upstream, "topic"))
	assert.Equal(t, []string{"topic"}, client.Topics())

	assert.NoError(t, br.Broker().Publish(newMessage("topic", "hello1")))
	assert.Equal(t, `hello1`, readStr(<-client.Channel()))

	assert.NoError(t, client.Unsubscribe("topic"))
	assert.NoError(t, client.Unsubscribe("topic"))
	assert.NoError(t, br.broker.Unsubscribe(client.upstream, "topic"))

	assert.Equal(t, []string{}, client.Topics())
}

func Test_Client_Patterns(t *testing.T) {
	br, _, c := setup(`{}`)
	defer c.Stop()

	client := br.NewClient()
	defer client.Close()

	assert.Equal(t, []string{}, client.Patterns())

	assert.NoError(t, client.SubscribePattern("topic/*"))
	assert.Equal(t, []string{"topic/*"}, client.Patterns())

	assert.NoError(t, br.broker.SubscribePattern(client.upstream, "topic/*"))
	assert.Equal(t, []string{"topic/*"}, client.Patterns())

	assert.NoError(t, br.Broker().Publish(newMessage("topic/1", "hello1")))
	assert.Equal(t, `hello1`, readStr(<-client.Channel()))

	assert.NoError(t, client.UnsubscribePattern("topic/*"))
	assert.NoError(t, br.broker.UnsubscribePattern(client.upstream, "topic/*"))

	assert.Equal(t, []string{}, client.Patterns())
}

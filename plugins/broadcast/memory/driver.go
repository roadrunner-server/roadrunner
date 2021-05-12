package memory

import "github.com/spiral/roadrunner/v2/plugins/broadcast"

type Driver struct {
}

func NewInMemoryDriver() broadcast.Subscriber {
	b := &Driver{}
	return b
}

func (d *Driver) Serve() error {
	panic("implement me")
}

func (d *Driver) Stop() {
	panic("implement me")
}

func (d *Driver) Subscribe(upstream chan *broadcast.Message, topics ...string) error {
	panic("implement me")
}

func (d *Driver) SubscribePattern(upstream chan *broadcast.Message, pattern string) error {
	panic("implement me")
}

func (d *Driver) Unsubscribe(upstream chan *broadcast.Message, topics ...string) error {
	panic("implement me")
}

func (d *Driver) UnsubscribePattern(upstream chan *broadcast.Message, pattern string) error {
	panic("implement me")
}

func (d *Driver) Publish(messages ...*broadcast.Message) error {
	panic("implement me")
}

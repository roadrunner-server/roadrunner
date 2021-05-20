package memory

import (
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
)

type Driver struct {
}

func NewInMemoryDriver() broadcast.Storage {
	b := &Driver{}
	return b
}

func (d *Driver) Store(uuid string, topics ...string) {
	panic("implement me")
}

func (d *Driver) StorePattern(uuid string, pattern string) {
	panic("implement me")
}

func (d *Driver) GetConnection(pattern string) []string {
	panic("implement me")
}

func (d *Driver) Construct(key string) (broadcast.Storage, error) {
	return nil, nil
}

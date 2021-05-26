package memory

import (
	"github.com/spiral/roadrunner/v2/plugins/memory/bst"
)

type Driver struct {
	tree bst.Storage
}

func NewInMemoryDriver() bst.Storage {
	b := &Driver{
		tree: bst.NewBST(),
	}
	return b
}

func (d *Driver) Insert(uuid string, topic string) {
	d.tree.Insert(uuid, topic)
}

func (d *Driver) Remove(uuid, topic string) {
	d.tree.Remove(uuid, topic)
}

func (d *Driver) Get(topic string) map[string]struct{} {
	return d.tree.Get(topic)
}

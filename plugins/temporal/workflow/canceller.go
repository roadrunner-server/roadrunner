package workflow

import (
	"sync"
)

type (
	cancellable func() error

	canceller struct {
		ids sync.Map
	}
)

func (c *canceller) register(id uint64, cancel cancellable) {
	c.ids.Store(id, cancel)
}

func (c *canceller) discard(id uint64) {
	c.ids.Delete(id)
}

func (c *canceller) cancel(ids ...uint64) error {
	var err error
	for _, id := range ids {
		cancel, ok := c.ids.LoadAndDelete(id)
		if ok == false {
			continue
		}

		err = cancel.(cancellable)()
		if err != nil {
			return err
		}
	}

	return nil
}

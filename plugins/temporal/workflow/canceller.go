package workflow

import (
	"sync"
)

type cancellable func() error

type canceller struct {
	ids sync.Map
}

func (c *canceller) register(id uint64, cancel cancellable) {
	c.ids.Store(id, cancel)
}

func (c *canceller) discard(id uint64) {
	c.ids.Delete(id)
}

func (c *canceller) cancel(ids ...uint64) error {
	var err error
	for _, id := range ids {
		cancel, ok := c.ids.Load(id)
		if ok == false {
			continue
		}

		// TODO return when minimum supported version will be go 1.15
		// go1.14 don't have LoadAndDelete method
		// It was introduced only in go1.15
		c.ids.Delete(id)

		err = cancel.(cancellable)()
		if err != nil {
			return err
		}
	}

	return nil
}

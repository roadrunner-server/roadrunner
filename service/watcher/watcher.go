package watcher

import (
	"log"
	"time"
)

// disconnect??
type Service struct {
	// defines how often
	interval time.Duration
	pool     Pool

	stop chan interface{}
}

// NewWatcher creates new pool watcher.
func NewWatcher(p Pool, i time.Duration) *Service {
	w := &Service{
		interval: i,
		pool:     p,
		stop:     make(chan interface{}),
	}

	go func() {
		ticker := time.NewTicker(w.interval)

		for {
			select {
			case <-ticker.C:
				w.update()
			case <-w.stop:
				return
			}
		}
	}()

	return w
}

func (w *Service) Stop() {
	close(w.stop)
}

func (w *Service) update() {
	for _, w := range w.pool.Workers() {
		log.Println(w)

	}
}

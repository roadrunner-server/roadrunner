package http

import "sync/atomic"

type statsListener struct {
	stats *ServiceStats
}

func (s *statsListener) listener(event int, ctx interface{}) {
	switch event {
	case EventResponse:
		atomic.AddUint64(&s.stats.Accepted, 1);
	case EventError:
		atomic.AddUint64(&s.stats.Accepted, 1);
	}
}

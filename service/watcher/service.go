package watcher

import (
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
)

// ID defines watcher service name.
const ID = "watch"

// Watchable defines the ability to attach roadrunner watcher.
type Watchable interface {
	// Watch attaches watcher to the service.
	Watch(w roadrunner.Watcher)
}

// Services to watch the state of roadrunner service inside other services.
type Service struct {
	cfg  *Config
	lsns []func(event int, ctx interface{})
}

// Init watcher service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	// mount Services to designated services
	for id, watcher := range cfg.Watchers(s.throw) {
		svc, _ := c.Get(id)
		if watchable, ok := svc.(Watchable); ok {
			watchable.Watch(watcher)
		}
	}

	return true, nil
}

// AddListener attaches server event watcher.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

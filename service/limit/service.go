package limit

import (
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
)

// ID defines controller service name.
const ID = "limit"

// controllable defines the ability to attach rr controller.
type controllable interface {
	// Attach attaches controller to the service.
	Attach(c roadrunner.Controller)
}

// Service to control the state of rr service inside other services.
type Service struct {
	lsns []func(event int, ctx interface{})
}

// Init controller service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	// mount Services to designated services
	for id, watcher := range cfg.Controllers(s.throw) {
		svc, _ := c.Get(id)
		if ctrl, ok := svc.(controllable); ok {
			ctrl.Attach(watcher)
		}
	}

	return true, nil
}

// AddListener attaches server event controller.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

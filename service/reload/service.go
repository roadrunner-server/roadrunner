package reload

import (
	"github.com/spiral/roadrunner/service"
	"os"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	reloadConfig *Config
}

// Init controller service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	// mount Services to designated services
	//for id, watcher := range cfg.Controllers(s.throw) {
	//	svc, _ := c.Get(id)
	//	if ctrl, ok := svc.(controllable); ok {
	//		ctrl.Attach(watcher)
	//	}
	//}

	s.reloadConfig = cfg

	return true, nil
}

func (s *Service) Serve() error {
	w, err := NewWatcher(SetMaxFileEvents(100))
	if err != nil {
		return err
	}

	name , _ := os.Getwd()

	w.AddSingle(name)

	println("test")

	return nil
}

func (s *Service) Stop() {

}
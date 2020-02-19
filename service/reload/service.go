package reload

import (
	"github.com/spiral/roadrunner/service"
	"os"
	"time"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	reloadConfig *Config
	container service.Container
}

// Init controller service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	s.container = c
	s.reloadConfig = cfg

	return true, nil
}

func (s *Service) Serve() error {
	w, err := NewWatcher(SetMaxFileEvents(100))
	if err != nil {
		return err
	}

	name , err := os.Getwd()
	if err != nil {
		return err
	}

	err = w.AddSingle(name)
	if err != nil {
		return err
	}

	go func() {
		err = w.StartPolling(time.Second)
		if err != nil {

		}
	}()



	// read events and restart corresponding services


	for {
		select {
		case e := <- w.Event:
			println(e.Name())
		}
	}
	//for e = range w.Event {
	//
	//	println("event")
	//	// todo use status
	//	//svc, _ := s.container.Get("http")
	//	//if svc != nil {
	//	//	if srv, ok := svc.(service.Service); ok {
	//	//		srv.Stop()
	//	//		err = srv.Serve()
	//	//		if err != nil {
	//	//			return err
	//	//		}
	//	//	}
	//	//}
	//
	//	//println("event skipped due to service is nil")
	//}


	return nil
}

func (s *Service) Stop() {

}
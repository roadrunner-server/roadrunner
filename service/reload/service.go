package reload

import (
	"github.com/spiral/roadrunner/service"
	"os"
	"strings"
	"time"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	reloadConfig *Config
	container    service.Container
	watcher      *Watcher
}

// Init controller service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	s.container = c
	s.reloadConfig = cfg

	return true, nil
}

func (s *Service) Serve() error {
	if !s.reloadConfig.Enabled {
		return nil
	}

	var err error
	s.watcher, err = NewWatcher([]WatcherConfig{WatcherConfig{
		serviceName: "test",
		recursive:   false,
		directories: []string{"/service"},
		filterHooks: func(filename, pattern string) error {
			if strings.Contains(filename, pattern) {
				return ErrorSkip
			}
			return nil
		},
		files:   make(map[string]os.FileInfo),
		//ignored: []string{".php"},
	}})
	if err != nil {
		return err
	}



	s.watcher.AddSingle("test", "/service")


	go func() {
		for {
			select {
			case e := <-s.watcher.Event:
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
	}()

	err = s.watcher.StartPolling(time.Second)
	if err != nil {
		return err
	}

	// read events and restart corresponding services

	return nil
}

func (s *Service) Stop() {
	//s.watcher.Stop()

}

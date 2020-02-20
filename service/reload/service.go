package reload

import (
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"os"
	"strings"
	"time"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	reloadConfig *Config
	watcher      *Watcher

}

// Init controller service
func (s *Service) Init(cfg *Config, c service.Container) (bool, error) {
	s.reloadConfig = cfg

	var err error
	var configs []WatcherConfig

	// mount Services to designated services
	for serviceName, _ := range cfg.Services {
		svc, _ := c.Get(serviceName)
		if ctrl, ok := svc.(roadrunner.Controllable); ok {
			tmp := cfg.Services[serviceName]
			tmp.service = &ctrl
			cfg.Services[serviceName] = tmp
		}
	}



	for serviceName, config := range s.reloadConfig.Services {
		if cfg.Services[serviceName].service == nil {
			continue
		}
		configs = append(configs, WatcherConfig{
			serviceName: serviceName,
			recursive:   config.Recursive,
			directories: config.Dirs,
			filterHooks: func(filename, pattern string) error {
				if strings.Contains(filename, pattern) {
					return ErrorSkip
				}
				return nil
			},
			files: make(map[string]os.FileInfo),
			//ignored:
		})
	}

	s.watcher, err = NewWatcher(configs)
	if err != nil {
		return false, err
	}

	for serviceName, config := range s.reloadConfig.Services {
		svc, _ := c.Get(serviceName)
		if ctrl, ok := svc.(*roadrunner.Controllable); ok {
			(*ctrl).Server().Reset()
		}

		configs = append(configs, WatcherConfig{
			serviceName: serviceName,
			recursive:   config.Recursive,
			directories: config.Dirs,
			filterHooks: func(filename, pattern string) error {
				if strings.Contains(filename, pattern) {
					return ErrorSkip
				}
				return nil
			},
			files: make(map[string]os.FileInfo),
			//ignored:

		})
	}

	return true, nil
}

func (s *Service) Serve() error {
	if !s.reloadConfig.Enabled {
		return nil
	}

	go func() {
		for {
			select {
			case e := <-s.watcher.Event:
				println(fmt.Sprintf("type is:%s, oldPath:%s, path:%s, name:%s", e.Type, e.OldPath, e.Path, e.FileInfo.Name()))

				srv := s.reloadConfig.Services[e.Type]

				if srv.service != nil {
					s := *srv.service
					err := s.Server().Reset()
					if err != nil {
						fmt.Println(err)
					}
				} else {
					s.watcher.mu.Lock()
					delete(s.watcher.watcherConfigs, e.Type)
					s.watcher.mu.Unlock()
				}
			}
		}
	}()

	err := s.watcher.StartPolling(time.Second * 2)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() {
	//s.watcher.Stop()

}

package reload

import (
	"errors"
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
	if !s.reloadConfig.Enabled {
		return false, nil
	}

	var configs []WatcherConfig

	// mount Services to designated services
	for serviceName := range cfg.Services {
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
		ignored, err := ConvertIgnored(config.Ignore)
		if err != nil {
			return false, err
		}
		configs = append(configs, WatcherConfig{
			serviceName: serviceName,
			recursive:   config.Recursive,
			directories: config.Dirs,
			filterHooks: func(filename string, patterns []string) error {
				for i := 0; i < len(patterns); i++ {
					if strings.Contains(filename, patterns[i]) {
						return nil
					}
				}
				return ErrorSkip
			},
			files:        make(map[string]os.FileInfo),
			ignored:      ignored,
			filePatterns: append(config.Patterns, cfg.Patterns...),
		})
	}

	var err error
	s.watcher, err = NewWatcher(configs)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) Serve() error {
	if !s.reloadConfig.Enabled {
		return nil
	}

	if s.reloadConfig.Interval < time.Second {
		return errors.New("reload interval is too fast")
	}

	go func() {
		for e := range s.watcher.Event {
			println(fmt.Sprintf("[UPDATE] Service: %s, path to file: %s, filename: %s", e.service, e.path, e.info.Name()))

			srv := s.reloadConfig.Services[e.service]

			if srv.service != nil {
				s := *srv.service
				err := s.Server().Reset()
				if err != nil {
					fmt.Println(err)
				}
			} else {
				s.watcher.mu.Lock()
				delete(s.watcher.watcherConfigs, e.service)
				s.watcher.mu.Unlock()
			}
		}
	}()

	err := s.watcher.StartPolling(s.reloadConfig.Interval)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() {
	s.watcher.Stop()
}

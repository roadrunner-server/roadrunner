package reload

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"os"
	"strings"
	"time"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	cfg     *Config
	log     *logrus.Logger
	watcher *Watcher
}

// Init controller service
func (s *Service) Init(cfg *Config, log *logrus.Logger, c service.Container) (bool, error) {
	if cfg == nil || len(cfg.Services) == 0 {
		return false, nil
	}

	s.cfg = cfg
	s.log = log

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

	for serviceName, config := range s.cfg.Services {
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
	if s.cfg.Interval < time.Second {
		return errors.New("reload interval is too fast")
	}

	go func() {
		for e := range s.watcher.Event {

			println(fmt.Sprintf("[UPDATE] Service: %s, path to file: %s, filename: %s", e.service, e.path, e.info.Name()))

			srv := s.cfg.Services[e.service]

			if srv.service != nil {
				sv := *srv.service
				err := sv.Server().Reset()
				if err != nil {
					s.log.Error(err)
				}
			} else {
				s.watcher.mu.Lock()
				delete(s.watcher.watcherConfigs, e.service)
				s.watcher.mu.Unlock()
			}
		}
	}()

	//go func() {
	//	for {
	//		select {
	//		case <-update:
	//			updated = append(updated, update)
	//		case <-time.Ticker:
	//			updated = updated[0:0]
	//			err := sv.Server().Reset()
	//			s.log.Debugf(
	//				"reload %s, found file changes in %s",
	//				strings.Join(updated, ","),
	//			)
	//		case <-exit:
	//		}
	//	}
	//}()

	err := s.watcher.StartPolling(s.cfg.Interval)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() {
	s.watcher.Stop()
}

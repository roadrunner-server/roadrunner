package reload

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
)

// ID contains default service name.
const ID = "reload"

type Service struct {
	cfg     *Config
	log     *logrus.Logger
	watcher *Watcher
	stopc   chan struct{}
}

// Init controller service
func (s *Service) Init(cfg *Config, log *logrus.Logger, c service.Container) (bool, error) {
	if cfg == nil || len(cfg.Services) == 0 {
		return false, nil
	}

	s.cfg = cfg
	s.log = log
	s.stopc = make(chan struct{})

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
			ServiceName: serviceName,
			Recursive:   config.Recursive,
			Directories: config.Dirs,
			FilterHooks: func(filename string, patterns []string) error {
				for i := 0; i < len(patterns); i++ {
					if strings.Contains(filename, patterns[i]) {
						return nil
					}
				}
				return ErrorSkip
			},
			Files:        make(map[string]os.FileInfo),
			Ignored:      ignored,
			FilePatterns: append(config.Patterns, cfg.Patterns...),
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

	// make a map with unique services
	// so, if we would have a 100 events from http service
	// in map we would see only 1 key and it's config
	treshholdc := make(chan struct {
		serviceConfig ServiceConfig
		service       string
	}, 100)

	// use the same interval
	timer := time.NewTimer(s.cfg.Interval)

	// drain channel in case of leaved messages
	defer func() {
		go func() {
			for range treshholdc {

			}
		}()
	}()

	go func() {
		for e := range s.watcher.Event {
			treshholdc <- struct {
				serviceConfig ServiceConfig
				service       string
			}{serviceConfig: s.cfg.Services[e.service], service: e.service}
		}
	}()

	// map with configs by services
	updated := make(map[string]ServiceConfig, 100)

	go func() {
		for {
			select {
			case config := <-treshholdc:
				// replace previous value in map by more recent without adding new one
				updated[config.service] = config.serviceConfig
				// stop timer
				timer.Stop()
				// restart
				// logic is following:
				// if we getting a lot of events, we should't restart particular service on each of it (user doing bug move or very fast typing)
				// instead, we are resetting the timer and wait for Interval time
				// If there is no more events, we restart service only once
				timer.Reset(s.cfg.Interval)
			case <-timer.C:
				if len(updated) > 0 {
					for k, v := range updated {
						sv := *v.service
						err := sv.Server().Reset()
						if err != nil {
							s.log.Error(err)
						}
						s.log.Debugf("[%s] found %v file(s) changes, reloading", k, len(updated))
					}
					// zero map
					updated = make(map[string]ServiceConfig, 100)
				}
			case <-s.stopc:
				timer.Stop()
				return
			}
		}
	}()

	err := s.watcher.StartPolling(s.cfg.Interval)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() {
	s.watcher.Stop()
	s.stopc <- struct{}{}
}

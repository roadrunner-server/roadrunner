package reload

import (
	"os"
	"strings"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// PluginName contains default plugin name.
const PluginName string = "reload"

type Service struct {
	cfg      *Config
	log      log.Logger
	watcher  *Watcher
	services map[string]interface{}
	stopc    chan struct{}
}

// Init controller service
func (s *Service) Init(cfg config.Configurer, log log.Logger) error {
	const op = errors.Op("reload plugin init")
	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		// disable plugin in case of error
		return errors.E(op, errors.Disabled, err)
	}

	s.log = log
	s.stopc = make(chan struct{})
	s.services = make(map[string]interface{})

	var configs []WatcherConfig

	for serviceName, serviceConfig := range s.cfg.Services {
		if s.cfg.Services[serviceName].service == nil {
			continue
		}
		ignored, err := ConvertIgnored(serviceConfig.Ignore)
		if err != nil {
			return errors.E(op, err)
		}
		configs = append(configs, WatcherConfig{
			serviceName: serviceName,
			recursive:   serviceConfig.Recursive,
			directories: serviceConfig.Dirs,
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
			filePatterns: append(serviceConfig.Patterns, s.cfg.Patterns...),
		})
	}

	s.watcher, err = NewWatcher(configs)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (s *Service) Serve() chan error {
	const op = errors.Op("reload plugin serve")
	errCh := make(chan error, 1)
	if s.cfg.Interval < time.Second {
		errCh <- errors.E(op, errors.Str("reload interval is too fast"))
		return errCh
	}

	// make a map with unique services
	// so, if we would have a 100 events from http service
	// in map we would see only 1 key and it's config
	treshholdc := make(chan struct {
		serviceConfig ServiceConfig
		service       string
	}, 100)

	// use the same interval
	ticker := time.NewTicker(s.cfg.Interval)

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
			case cfg := <-treshholdc:
				// replace previous value in map by more recent without adding new one
				updated[cfg.service] = cfg.serviceConfig
				// stop ticker
				ticker.Stop()
				// restart
				// logic is following:
				// if we getting a lot of events, we should't restart particular service on each of it (user doing bug move or very fast typing)
				// instead, we are resetting the ticker and wait for Interval time
				// If there is no more events, we restart service only once
				ticker = time.NewTicker(s.cfg.Interval)
			case <-ticker.C:
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
				ticker.Stop()
				return
			}
		}
	}()

	err := s.watcher.StartPolling(s.cfg.Interval)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	return errCh
}

func (s *Service) Stop() {
	s.watcher.Stop()
	s.stopc <- struct{}{}
}

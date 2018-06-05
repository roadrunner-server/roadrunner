package _____

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/rpc"
	"sync"
)

// Config provides ability to slice configuration sections and unmarshal configuration data into
// given structure.
type Config interface {
	// Get nested config section (sub-map), returns nil if section not found.
	Get(service string) Config

	// Unmarshal unmarshal config data into given struct.
	Unmarshal(out interface{}) error
}

var (
	dsnError = errors.New("invalid socket DSN (tcp://:6001, unix://sock.unix)")
)

type Bus struct {
	services  []Service
	wg        sync.WaitGroup
	enabled   []Service
	stop      chan interface{}
	rpc       *rpc.Server
	rpcConfig *RPCConfig
}

func (b *Bus) Register(s Service) {
	b.services = append(b.services, s)
}

func (b *Bus) Services() []Service {
	return b.services
}

func (b *Bus) Configure(cfg Config) error {
	b.enabled = make([]Service, 0)

	for _, s := range b.services {
		segment := cfg.Get(s.Name())
		if segment == nil {
			// no config has been provided for the Service
			logrus.Debugf("%s: no config has been provided", s.Name())
			continue
		}

		if enable, err := s.Configure(segment); err != nil {
			return err
		} else if !enable {
			continue
		}

		b.enabled = append(b.enabled, s)
	}

	return nil
}

func (b *Bus) Serve() {
	b.rpc = rpc.NewServer()

	for _, s := range b.enabled {
		// some candidates might provide net/rpc api for internal communications
		if api := s.RPC(); api != nil {
			b.rpc.RegisterName(s.Name(), api)
		}

		b.wg.Add(1)
		go func() {
			defer b.wg.Done()

			if err := s.Serve(); err != nil {
				logrus.Errorf("%s.start: %s", s.Name(), err)
			}
		}()
	}

	b.wg.Wait()
}

func (b *Bus) Stop() {
	for _, s := range b.enabled {
		if err := s.Stop(); err != nil {
			logrus.Errorf("%s.stop: %s", s.Name(), err)
		}
	}

	b.wg.Wait()
}

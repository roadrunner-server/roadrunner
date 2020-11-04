package rpc

import (
	"net/rpc"

	"go.uber.org/zap"

	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// Pluggable declares the ability to create set of public RPC methods.
type Pluggable interface {
	endure.Named

	// Provides RPC methods for the given service.
	RPCService() (interface{}, error)
}

// ServiceName contains default service name.
const ServiceName = "rpc"

// Plugin is RPC service.
type Plugin struct {
	cfg      Config
	log      *zap.Logger
	rpc      *rpc.Server
	services []Pluggable
	close    chan struct{}
}

// Init rpc service. Must return true if service is enabled.
func (s *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("RPC Init")
	if !cfg.Has(ServiceName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(ServiceName, &s.cfg)
	if err != nil {
		return err
	}
	s.cfg.InitDefaults()

	if s.cfg.Disabled {
		return errors.E(op, errors.Disabled)
	}

	s.log = log
	s.close = make(chan struct{}, 1)

	return s.cfg.Valid()
}

// Serve serves the service.
func (s *Plugin) Serve() chan error {
	const op = errors.Op("register service")
	errCh := make(chan error, 1)

	s.rpc = rpc.NewServer()

	services := make([]string, 0, len(s.services))

	// Attach all services
	for i := 0; i < len(s.services); i++ {
		svc, err := s.services[i].RPCService()
		if err != nil {
			errCh <- errors.E(op, err)
			return errCh
		}

		err = s.Register(s.services[i].Name(), svc)
		if err != nil {
			errCh <- errors.E(op, err)
			return errCh
		}

		services = append(services, s.services[i].Name())
	}

	ln, err := s.cfg.Listener()
	if err != nil {
		errCh <- err
		return errCh
	}

	s.log.Debug("Started RPC service", zap.String("address", s.cfg.Listen), zap.Any("services", services))

	go func() {
		for {
			select {
			case <-s.close:
				// log error
				err = ln.Close()
				if err != nil {
					errCh <- errors.E(errors.Op("close RPC socket"), err)
				}
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					continue
				}

				go s.rpc.ServeCodec(goridge.NewCodec(conn))
			}
		}
	}()

	return errCh
}

// Stop stops the service.
func (s *Plugin) Stop() error {
	s.close <- struct{}{}
	return nil
}

// Name contains service name.
func (s *Plugin) Name() string {
	return ServiceName
}

// Depends declares services to collect for RPC.
func (s *Plugin) Collects() []interface{} {
	return []interface{}{
		s.RegisterPlugin,
	}
}

// RegisterPlugin registers RPC service plugin.
func (s *Plugin) RegisterPlugin(p Pluggable) error {
	s.services = append(s.services, p)
	return nil
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
func (s *Plugin) Register(name string, svc interface{}) error {
	if s.rpc == nil {
		return errors.E("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, svc)
}

// Client creates new RPC client.
func (s *Plugin) Client() (*rpc.Client, error) {
	conn, err := s.cfg.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridge.NewClientCodec(conn)), nil
}

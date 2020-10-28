package rpc

import (
	"net/rpc"

	"go.uber.org/zap"

	"github.com/spiral/endure"
	"github.com/spiral/endure/errors"
	"github.com/spiral/goridge/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// RPCPluggable declares the ability to create set of public RPC methods.
type RPCPluggable interface {
	endure.Named

	// Provides RPC methods for the given service.
	RPCService() (interface{}, error)
}

// ServiceName contains default service name.
const ServiceName = "rpc"

// Service is RPC service.
type Service struct {
	cfg      Config
	log      *zap.Logger
	rpc      *rpc.Server
	services []RPCPluggable
	close    chan struct{}
}

// Init rpc service. Must return true if service is enabled.
func (s *Service) Init(cfg config.Provider, log *zap.Logger) error {
	if !cfg.Has(ServiceName) {
		return errors.E(errors.Disabled)
	}

	err := cfg.UnmarshalKey(ServiceName, &s.cfg)
	if err != nil {
		return err
	}
	s.cfg.InitDefaults()

	if s.cfg.Disabled {
		return errors.E(errors.Disabled)
	}

	s.log = log

	return s.cfg.Valid()
}

// Serve serves the service.
func (s *Service) Serve() chan error {
	errCh := make(chan error, 1)

	s.close = make(chan struct{}, 1)
	s.rpc = rpc.NewServer()

	names := make([]string, 0, len(s.services))

	// Attach all services
	for i := 0; i < len(s.services); i++ {
		svc, err := s.services[i].RPCService()
		if err != nil {
			errCh <- errors.E(errors.Op("register service"), err)
			return errCh
		}

		err = s.Register(s.services[i].Name(), svc)
		if err != nil {
			errCh <- errors.E(errors.Op("register service"), err)
			return errCh
		}

		names = append(names, s.services[i].Name())
	}

	ln, err := s.cfg.Listener()
	if err != nil {
		errCh <- err
		return errCh
	}

	s.log.Debug("Started RPC service", zap.String("address", s.cfg.Listen), zap.Any("services", names))

	go func() {
		for {
			select {
			case <-s.close:
				// log error
				err := ln.Close()
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
func (s *Service) Stop() error {
	s.close <- struct{}{}
	return nil
}

// Name contains service name.
func (s *Service) Name() string {
	return ServiceName
}

// Depends declares services to collect for RPC.
func (s *Service) Depends() []interface{} {
	return []interface{}{
		s.RegisterPlugin,
	}
}

// RegisterPlugin registers RPC service plugin.
func (s *Service) RegisterPlugin(p RPCPluggable) error {
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
func (s *Service) Register(name string, svc interface{}) error {
	if s.rpc == nil {
		return errors.E("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, svc)
}

// Client creates new RPC client.
func (s *Service) Client() (*rpc.Client, error) {
	conn, err := s.cfg.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridge.NewClientCodec(conn)), nil
}

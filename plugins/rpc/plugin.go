package rpc

import (
	"net"
	"net/rpc"
	"sync/atomic"

	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v2"
	"github.com/spiral/roadrunner/v2/log"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

// Pluggable declares the ability to create set of public RPC methods.
type Pluggable interface {
	endure.Named

	// Provides RPC methods for the given service.
	RPCService() (interface{}, error)
}

// ServiceName contains default service name.
const ServiceName = "RPC"

// Plugin is RPC service.
type Plugin struct {
	cfg      Config
	log      log.Logger
	rpc      *rpc.Server
	services []Pluggable
	listener net.Listener
	closed   *uint32
}

// Init rpc service. Must return true if service is enabled.
func (s *Plugin) Init(cfg config.Configurer, log log.Logger) error {
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
	state := uint32(0)
	s.closed = &state
	atomic.StoreUint32(s.closed, 0)

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

	var err error
	s.listener, err = s.cfg.Listener()
	if err != nil {
		errCh <- err
		return errCh
	}

	s.log.Debug("Started RPC service", "address", s.cfg.Listen, "services", services)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if atomic.LoadUint32(s.closed) == 1 {
					// just log and continue, this is not a critical issue, we just called Stop
					s.log.Error("listener accept error, connection closed", "error", err)
					return
				}

				s.log.Error("listener accept error", "error", err)
				errCh <- errors.E(errors.Op("listener accept"), errors.Serve, err)
				return
			}

			go s.rpc.ServeCodec(goridge.NewCodec(conn))
		}
	}()

	return errCh
}

// Stop stops the service.
func (s *Plugin) Stop() error {
	// store closed state
	atomic.StoreUint32(s.closed, 1)
	err := s.listener.Close()
	if err != nil {
		return errors.E(errors.Op("stop RPC socket"), err)
	}
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

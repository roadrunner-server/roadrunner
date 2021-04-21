package rpc

import (
	"net"
	"net/rpc"
	"sync/atomic"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// PluginName contains default plugin name.
const PluginName = "rpc"

// Plugin is RPC service.
type Plugin struct {
	cfg Config
	log logger.Logger
	rpc *rpc.Server
	// set of the plugins, which are implement RPCer interface and can be plugged into the RR via RPC
	plugins  map[string]RPCer
	listener net.Listener
	closed   uint32
}

// Init rpc service. Must return true if service is enabled.
func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("rpc_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}
	// Init defaults
	s.cfg.InitDefaults()
	// Init pluggable plugins map
	s.plugins = make(map[string]RPCer, 5)
	// init logs
	s.log = log

	// set up state
	atomic.StoreUint32(&s.closed, 0)

	// validate config
	err = s.cfg.Valid()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Serve serves the service.
func (s *Plugin) Serve() chan error {
	const op = errors.Op("rpc_plugin_serve")
	errCh := make(chan error, 1)

	s.rpc = rpc.NewServer()

	services := make([]string, 0, len(s.plugins))

	// Attach all services
	for name := range s.plugins {
		err := s.Register(name, s.plugins[name].RPC())
		if err != nil {
			errCh <- errors.E(op, err)
			return errCh
		}

		services = append(services, name)
	}

	var err error
	s.listener, err = s.cfg.Listener()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	s.log.Debug("Started RPC service", "address", s.cfg.Listen, "services", services)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				if atomic.LoadUint32(&s.closed) == 1 {
					// just continue, this is not a critical issue, we just called Stop
					return
				}

				s.log.Error("listener accept error", "error", err)
				errCh <- errors.E(errors.Op("listener accept"), errors.Serve, err)
				return
			}

			go s.rpc.ServeCodec(goridgeRpc.NewCodec(conn))
		}
	}()

	return errCh
}

// Stop stops the service.
func (s *Plugin) Stop() error {
	const op = errors.Op("rpc_plugin_stop")
	// store closed state
	atomic.StoreUint32(&s.closed, 1)
	err := s.listener.Close()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Name contains service name.
func (s *Plugin) Name() string {
	return PluginName
}

// Collects all plugins which implement Name + RPCer interfaces
func (s *Plugin) Collects() []interface{} {
	return []interface{}{
		s.RegisterPlugin,
	}
}

// RegisterPlugin registers RPC service plugin.
func (s *Plugin) RegisterPlugin(name endure.Named, p RPCer) {
	s.plugins[name.Name()] = p
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

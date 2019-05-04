package rpc

import (
	"errors"
	"github.com/spiral/goridge"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/env"
	"net/rpc"
	"sync"
)

// ID contains default service name.
const ID = "rpc"

// Service is RPC service.
type Service struct {
	cfg     *Config
	stop    chan interface{}
	rpc     *rpc.Server
	mu      sync.Mutex
	serving bool
}

// Init rpc service. Must return true if service is enabled.
func (s *Service) Init(cfg *Config, c service.Container, env env.Environment) (bool, error) {
	if !cfg.Enable {
		return false, nil
	}

	s.cfg = cfg
	s.rpc = rpc.NewServer()

	if env != nil {
		env.SetEnv("RR_RPC", cfg.Listen)
	}

	if err := s.Register("system", &systemService{c}); err != nil {
		return false, err
	}

	return true, nil
}

// Serve serves the service.
func (s *Service) Serve() error {
	if s.rpc == nil {
		return errors.New("RPC service is not configured")
	}

	s.mu.Lock()
	s.serving = true
	s.stop = make(chan interface{})
	s.mu.Unlock()

	ln, err := s.cfg.Listener()
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() {
		for {
			select {
			case <-s.stop:
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

	<-s.stop

	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	return nil
}

// Stop stops the service.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serving {
		close(s.stop)
	}
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
		return errors.New("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, svc)
}

// Client creates new RPC client.
func (s *Service) Client() (*rpc.Client, error) {
	if s.cfg == nil {
		return nil, errors.New("RPC service is not configured")
	}

	conn, err := s.cfg.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridge.NewClientCodec(conn)), nil
}

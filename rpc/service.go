package rpc

import (
	"errors"
	"github.com/spiral/goridge"
	"github.com/spiral/roadrunner/service"
	"net/rpc"
	"sync"
)

// Service is RPC service.
type Service struct {
	cfg  *config
	stop chan interface{}
	rpc  *rpc.Server

	mu      sync.Mutex
	serving bool
}

// WithConfig must return Service instance configured with the given environment. Must return error in case of
// misconfiguration, might return nil as Service if Service is not enabled.
func (s *Service) WithConfig(cfg service.Config, reg service.Registry) (service.Service, error) {
	config := &config{}
	if err := cfg.Unmarshal(config); err != nil {
		return nil, err
	}

	if !config.Enable {
		return nil, nil
	}

	return &Service{cfg: config, rpc: rpc.NewServer()}, nil
}

// Serve serves Service.
func (s *Service) Serve() error {
	if s.rpc == nil {
		return errors.New("RPC service is not configured")
	}

	s.mu.Lock()
	s.serving = true
	s.stop = make(chan interface{})
	s.mu.Unlock()

	ln, err := s.cfg.listener()
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		select {
		case <-s.stop:
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}

			s.rpc.Accept(ln)
			go s.rpc.ServeCodec(goridge.NewCodec(conn))
		}
	}

	return nil
}

// Close stop Service Service.
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
func (s *Service) Register(name string, rcvr interface{}) error {
	if s.rpc == nil {
		return errors.New("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, rcvr)
}

// Client creates new RPC client.
func (s *Service) Client() (*rpc.Client, error) {
	if s.cfg == nil {
		return nil, errors.New("RPC service is not configured")
	}

	conn, err := s.cfg.dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridge.NewClientCodec(conn)), nil
}

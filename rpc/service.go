package rpc

import (
	"errors"
	"github.com/spiral/goridge"
	"github.com/spiral/roadrunner/service"
	"net/rpc"
	"sync"
)

// todo: improved logging

// Service is RPC service.
type Service struct {
	cfg  *config
	stop chan interface{}
	rpc  *rpc.Server

	mu      sync.Mutex
	serving bool
}

// Configure must return configure service and return true if service is enabled. Must return error in case of
// misconfiguration.
func (s *Service) Configure(cfg service.Config, reg service.Container) (enabled bool, err error) {
	config := &config{}
	if err := cfg.Unmarshal(config); err != nil {
		return false, err
	}

	if !config.Enable {
		return false, nil
	}

	s.cfg = config
	s.rpc = rpc.NewServer()

	return true, nil
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

	go func() {
		for {
			select {
			case <-s.stop:
				break
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

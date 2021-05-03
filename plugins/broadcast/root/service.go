package broadcast

import (
	"errors"
	"sync"

	"github.com/spiral/roadrunner/service/rpc"
)

// ID defines public service name.
const ID = "broadcast"

// Service manages even broadcasting and websocket interface.
type Service struct {
	// service and broker configuration
	cfg *Config

	// broker
	mu     sync.Mutex
	broker Broker
}

// Init service.
func (s *Service) Init(cfg *Config, rpc *rpc.Service) (ok bool, err error) {
	s.cfg = cfg

	if rpc != nil {
		if err := rpc.Register(ID, &rpcService{svc: s}); err != nil {
			return false, err
		}
	}

	s.mu.Lock()
	if s.cfg.Redis != nil {
		if s.broker, err = redisBroker(s.cfg.Redis); err != nil {
			return false, err
		}
	} else {
		s.broker = memoryBroker()
	}
	s.mu.Unlock()

	return true, nil
}

// Serve broadcast broker.
func (s *Service) Serve() (err error) {
	return s.broker.Serve()
}

// Stop closes broadcast broker.
func (s *Service) Stop() {
	broker := s.Broker()
	if broker != nil {
		broker.Stop()
	}
}

// Broker returns associated broker.
func (s *Service) Broker() Broker {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.broker
}

// NewClient returns single connected client with ability to consume or produce into associated topic(svc).
func (s *Service) NewClient() *Client {
	return &Client{
		upstream: make(chan *Message),
		broker:   s.Broker(),
		topics:   make([]string, 0),
		patterns: make([]string, 0),
	}
}

// Publish one or multiple Channel.
func (s *Service) Publish(msg ...*Message) error {
	broker := s.Broker()
	if broker == nil {
		return errors.New("no stopped broker")
	}

	return s.Broker().Publish(msg...)
}

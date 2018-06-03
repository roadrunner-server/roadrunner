package service

import (
	"github.com/sirupsen/logrus"
	"net/rpc"
	"sync"
	"github.com/spiral/goridge"
	"github.com/pkg/errors"
)

const (
	rpcConfig = "rpc"
)

var (
	dsnError = errors.New("invalid socket DSN (tcp://:6001, unix://sock.unix)")
)

type Bus struct {
	wg        sync.WaitGroup
	services  []Service
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
	if segment := cfg.Get(rpcConfig); segment == nil {
		logrus.Warn("rpc: no config has been provided")
	} else {
		b.rpcConfig = &RPCConfig{}
		if err := segment.Unmarshal(b.rpcConfig); err != nil {
			return err
		}
	}

	b.enabled = make([]Service, 0)

	for _, s := range b.services {
		segment := cfg.Get(s.Name())
		if segment == nil {
			// no config has been provided for the service
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

func (b *Bus) RCPClient() (*rpc.Client, error) {
	if b.rpcConfig == nil {
		return nil, errors.New("rpc is not configured")
	}

	conn, err := b.rpcConfig.CreateDialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridge.NewClientCodec(conn)), nil
}

func (b *Bus) Serve() {
	b.rpc = rpc.NewServer()

	for _, s := range b.enabled {
		// some services might provide net/rpc api for internal communications
		if api := s.RPC(); api != nil {
			b.rpc.RegisterName(s.Name(), api)
		}

		b.wg.Add(1)
		go func() {
			defer b.wg.Done()

			if err := s.Serve(); err != nil {
				logrus.Errorf("%s.start: ", s.Name(), err)
			}
		}()
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()

		logrus.Debug("rpc: started")
		if err := b.serveRPC(); err != nil {
			logrus.Errorf("rpc: %s", err)
		}
	}()

	b.wg.Wait()
}

func (b *Bus) Stop() {
	if err := b.stopRPC(); err != nil {
		logrus.Errorf("rpc: ", err)
	}

	for _, s := range b.enabled {
		if err := s.Stop(); err != nil {
			logrus.Errorf("%s.stop: ", s.Name(), err)
		}
	}

	b.wg.Wait()
}

func (b *Bus) serveRPC() error {
	if b.rpcConfig == nil {
		return nil
	}

	b.stop = make(chan interface{})

	ln, err := b.rpcConfig.CreateListener()
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		select {
		case <-b.stop:
			b.rpc = nil
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}

			go b.rpc.ServeCodec(goridge.NewCodec(conn))
		}
	}

	return nil
}

func (b *Bus) stopRPC() error {
	if b.rpcConfig == nil {
		return nil
	}

	close(b.stop)
	return nil
}

package service

import (
	"net"
	"strings"
)

type RPCConfig struct {
	Listen string
}

func (cfg *RPCConfig) CreateListener() (net.Listener, error) {
	dsn := strings.Split(cfg.Listen, "://")
	if len(dsn) != 2 {
		return nil, dsnError
	}

	return net.Listen(dsn[0], dsn[1])
}

func (cfg *RPCConfig) CreateDialer() (net.Conn, error) {
	dsn := strings.Split(cfg.Listen, "://")
	if len(dsn) != 2 {
		return nil, dsnError
	}

	return net.Dial(dsn[0], dsn[1])
}

func NewBus() *Bus {
	return &Bus{services: make([]Service, 0)}
}

package grpc

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

const (
	name string = "grpc"
)

type Plugin struct {
	config *Config
	opts   []grpc.ServerOption

	cfg config.Configurer
	log logger.Logger
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("grpc_plugin_init")

	// register the codec
	encoding.RegisterCodec(&codec{})

	return nil
}

func (p *Plugin) Serve() chan error {
	const op = errors.Op("grpc_plugin_serve")
	errCh := make(chan error, 1)

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Available() {}

func (p *Plugin) Name() string {
	return name
}

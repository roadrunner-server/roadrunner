package grpc

import "github.com/spiral/errors"

type Plugin struct {
}

func (p *Plugin) Init() error {
	const op = errors.Op("grpc_plugin_init")
	return nil
}

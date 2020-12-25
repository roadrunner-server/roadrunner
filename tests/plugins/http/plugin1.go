package http

import (
	"github.com/spiral/roadrunner/v2/plugins/config"
)

type Plugin1 struct {
	config config.Configurer
}

func (p1 *Plugin1) Init(cfg config.Configurer) error {
	p1.config = cfg
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Name() string {
	return "http_test.plugin1"
}

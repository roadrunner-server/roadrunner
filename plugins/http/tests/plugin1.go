package tests

import (
	config2 "github.com/spiral/roadrunner/v2/interfaces/config"
)

type Plugin1 struct {
	config config2.Configurer
}

func (p1 *Plugin1) Init(cfg config2.Configurer) error {
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

package config

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

type Foo2 struct {
	configProvider config.Configurer
}

// Depends on S2 and DB (S3 in the current case)
func (f *Foo2) Init(p config.Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo2) Serve() chan error {
	const op = errors.Op("foo_plugin_serve")
	errCh := make(chan error, 1)

	r := &ReloadConfig{}
	err := f.configProvider.UnmarshalKey("reload", r)
	if err != nil {
		errCh <- err
	}

	if len(r.Patterns) == 0 {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	var allCfg AllConfig
	err = f.configProvider.Unmarshal(&allCfg)
	if err != nil {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	if allCfg.RPC.Listen != "tcp://localhost:6061" {
		errCh <- errors.E(op, errors.Str("RPC.Listen should be overwritten"))
		return errCh
	}

	return errCh
}

func (f *Foo2) Stop() error {
	return nil
}

package tests

import (
	"errors"
	"fmt"

	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/factory"
)

type Foo struct {
	configProvider config.Provider
	spawner        factory.Spawner
}

func (f *Foo) Init(p config.Provider, spw factory.Spawner) error {
	f.configProvider = p
	f.spawner = spw
	return nil
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error, 1)

	r := &factory.Config{}
	err := f.configProvider.UnmarshalKey("app", r)
	if err != nil {
		errCh <- err
		return errCh
	}

	cmd, err := f.spawner.CommandFactory(nil)
	if err != nil {
		errCh <- err
		return errCh
	}
	if cmd == nil {
		errCh <- errors.New("command is nil")
		return errCh
	}
	a := cmd()
	out, err := a.Output()
	if err != nil {
		errCh <- err
		return errCh
	}

	fmt.Println(string(out))

	return errCh
}

func (f *Foo) Stop() error {
	return nil
}

package tests

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/temporalio/roadrunner-temporal/config"
	"github.com/temporalio/roadrunner-temporal/factory"
	"github.com/temporalio/roadrunner-temporal/roadrunner"
)

type Foo2 struct {
	configProvider config.Provider
	wf             factory.WorkerFactory
	spw            factory.Spawner
}

func (f *Foo2) Init(p config.Provider, workerFactory factory.WorkerFactory, spawner factory.Spawner) error {
	f.configProvider = p
	f.wf = workerFactory
	f.spw = spawner
	return nil
}

func (f *Foo2) Serve() chan error {
	errCh := make(chan error, 1)

	r := &factory.AppConfig{}
	err := f.configProvider.UnmarshalKey("app", r)
	if err != nil {
		errCh <- err
		return errCh
	}

	cmd, err := f.spw.NewCmd(nil)
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

	w, err := f.wf.NewWorker(context.Background(), nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	_ = w

	poolConfig := &roadrunner.Config{
		NumWorkers:      10,
		MaxJobs:         100,
		AllocateTimeout: time.Second * 10,
		DestroyTimeout:  time.Second * 10,
		TTL:             1000,
		IdleTTL:         1000,
		ExecTTL:         time.Second * 10,
		MaxPoolMemory:   10000,
		MaxWorkerMemory: 10000,
	}

	pool, err := f.wf.NewWorkerPool(context.Background(), poolConfig, nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	_ = pool

	fmt.Println(string(out))

	return errCh
}

func (f *Foo2) Stop() error {
	return nil
}

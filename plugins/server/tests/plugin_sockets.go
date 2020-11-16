package tests

import (
	"context"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/interfaces/server"
	"github.com/spiral/roadrunner/v2/plugins/config"
	plugin "github.com/spiral/roadrunner/v2/plugins/server"
)

type Foo2 struct {
	configProvider config.Configurer
	wf             server.WorkerFactory
	pool           roadrunner.Pool
}

func (f *Foo2) Init(p config.Configurer, workerFactory server.WorkerFactory) error {
	f.configProvider = p
	f.wf = workerFactory
	return nil
}

func (f *Foo2) Serve() chan error {
	const op = errors.Op("serve")
	var err error
	errCh := make(chan error, 1)
	conf := &plugin.Config{}

	// test payload for echo
	r := roadrunner.Payload{
		Context: nil,
		Body:    []byte(Response),
	}

	err = f.configProvider.UnmarshalKey(ConfigSection, conf)
	if err != nil {
		errCh <- err
		return errCh
	}

	// test CMDFactory
	cmd, err := f.wf.CmdFactory(nil)
	if err != nil {
		errCh <- err
		return errCh
	}
	if cmd == nil {
		errCh <- errors.E(op, "command is nil")
		return errCh
	}

	// test worker creation
	w, err := f.wf.NewWorker(context.Background(), nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	// test that our worker is functional
	sw, err := roadrunner.NewSyncWorker(w)
	if err != nil {
		errCh <- err
		return errCh
	}

	rsp, err := sw.Exec(r)
	if err != nil {
		errCh <- err
		return errCh
	}

	if string(rsp.Body) != Response {
		errCh <- errors.E("response from worker is wrong", errors.Errorf("response: %s", rsp.Body))
		return errCh
	}

	// should not be errors
	err = sw.Stop(context.Background())
	if err != nil {
		errCh <- err
		return errCh
	}

	// test pool
	f.pool, err = f.wf.NewWorkerPool(context.Background(), testPoolConfig, nil)
	if err != nil {
		errCh <- err
		return errCh
	}

	// test pool execution
	rsp, err = f.pool.Exec(r)
	if err != nil {
		errCh <- err
		return errCh
	}

	// echo of the "test" should be -> test
	if string(rsp.Body) != Response {
		errCh <- errors.E("response from worker is wrong", errors.Errorf("response: %s", rsp.Body))
		return errCh
	}

	return errCh
}

func (f *Foo2) Stop() error {
	f.pool.Destroy(context.Background())
	return nil
}

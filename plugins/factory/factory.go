package factory

import (
	"context"

	"log"

	"github.com/fatih/color"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/plugins/events"
)

type WorkerFactory interface {
	NewWorker(ctx context.Context, env Env) (roadrunner.WorkerBase, error)
	NewWorkerPool(ctx context.Context, opt *roadrunner.Config, env Env) (roadrunner.Pool, error)
}

type WFactory struct {
	events   *events.EventBroadcaster
	app      Spawner
	wFactory roadrunner.Factory
}

func (wf *WFactory) Init(app Spawner) (err error) {
	wf.events = events.NewEventBroadcaster()

	wf.app = app
	wf.wFactory, err = app.NewFactory()
	if err != nil {
		return nil
	}

	return nil
}

func (wf *WFactory) AddListener(l events.EventListener) {
	wf.events.AddListener(l)
}

func (wf *WFactory) NewWorkerPool(ctx context.Context, opt *roadrunner.Config, env Env) (roadrunner.Pool, error) {
	cmd, err := wf.app.NewCmd(env)
	if err != nil {
		return nil, err
	}

	p, err := roadrunner.NewPool(ctx, cmd, wf.wFactory, opt)
	if err != nil {
		return nil, err
	}

	// TODO event to stop
	go func() {
		for e := range p.Events() {
			wf.events.Push(e)
			if we, ok := e.Payload.(roadrunner.WorkerEvent); ok {
				if we.Event == roadrunner.EventWorkerLog {
					log.Print(color.YellowString(string(we.Payload.([]byte))))
				}
			}
		}
	}()

	return p, nil
}

func (wf *WFactory) NewWorker(ctx context.Context, env Env) (roadrunner.WorkerBase, error) {
	cmd, err := wf.app.NewCmd(env)
	if err != nil {
		return nil, err
	}

	return wf.wFactory.SpawnWorkerWithContext(ctx, cmd())
}

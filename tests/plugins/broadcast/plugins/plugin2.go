package plugins

import (
	"context"
	"fmt"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/common/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const Plugin2Name = "plugin2"

type Plugin2 struct {
	log    logger.Logger
	b      broadcast.Broadcaster
	driver pubsub.SubReader
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *Plugin2) Init(log logger.Logger, b broadcast.Broadcaster) error {
	p.log = log
	p.b = b
	p.ctx, p.cancel = context.WithCancel(context.Background())
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test")
	if err != nil {
		panic(err)
	}

	err = p.driver.Subscribe("2", "foo")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, err := p.driver.Next(p.ctx)
			if err != nil {
				if errors.Is(errors.TimeOut, err) {
					return
				}
				errCh <- err
				return
			}

			if msg == nil {
				continue
			}

			p.log.Info(fmt.Sprintf("%s: %s", Plugin2Name, *msg))
		}
	}()

	return errCh
}

func (p *Plugin2) Stop() error {
	_ = p.driver.Unsubscribe("2", "foo")
	p.cancel()
	return nil
}

func (p *Plugin2) Name() string {
	return Plugin2Name
}

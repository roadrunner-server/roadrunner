package plugins

import (
	"fmt"

	"github.com/spiral/roadrunner/v2/common/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const Plugin1Name = "plugin1"

type Plugin1 struct {
	log    logger.Logger
	b      broadcast.Broadcaster
	driver pubsub.SubReader

	exit chan struct{}
}

func (p *Plugin1) Init(log logger.Logger, b broadcast.Broadcaster) error {
	p.log = log
	p.b = b
	p.exit = make(chan struct{}, 1)
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test")
	if err != nil {
		errCh <- err
		return errCh
	}

	err = p.driver.Subscribe("1", "foo", "foo2", "foo3")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case <-p.exit:
				return
			default:
				msg, err := p.driver.Next()
				if err != nil {
					errCh <- err
					return
				}

				if msg == nil {
					continue
				}

				p.log.Info(fmt.Sprintf("%s: %s", Plugin1Name, *msg))
			}
		}
	}()

	return errCh
}

func (p *Plugin1) Stop() error {
	_ = p.driver.Unsubscribe("1", "foo")
	_ = p.driver.Unsubscribe("1", "foo2")
	_ = p.driver.Unsubscribe("1", "foo3")

	p.exit <- struct{}{}
	return nil
}

func (p *Plugin1) Name() string {
	return Plugin1Name
}

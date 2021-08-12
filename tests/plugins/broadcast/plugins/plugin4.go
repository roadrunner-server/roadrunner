package plugins

import (
	"fmt"

	"github.com/spiral/roadrunner/v2/common/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const Plugin4Name = "plugin4"

type Plugin4 struct {
	log    logger.Logger
	b      broadcast.Broadcaster
	driver pubsub.SubReader

	exit chan struct{}
}

func (p *Plugin4) Init(log logger.Logger, b broadcast.Broadcaster) error {
	p.log = log
	p.b = b

	p.exit = make(chan struct{}, 1)
	return nil
}

func (p *Plugin4) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test3")
	if err != nil {
		panic(err)
	}

	err = p.driver.Subscribe("4", "foo")
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

				p.log.Info(fmt.Sprintf("%s: %s", Plugin4Name, *msg))
			}
		}
	}()

	return errCh
}

func (p *Plugin4) Stop() error {
	_ = p.driver.Unsubscribe("4", "foo")

	p.exit <- struct{}{}
	return nil
}

func (p *Plugin4) Name() string {
	return Plugin4Name
}

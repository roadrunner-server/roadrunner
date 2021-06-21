package plugins

import (
	"fmt"

	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const Plugin3Name = "plugin3"

type Plugin3 struct {
	log    logger.Logger
	b      broadcast.Broadcaster
	driver pubsub.SubReader
}

func (p *Plugin3) Init(log logger.Logger, b broadcast.Broadcaster) error {
	p.log = log
	p.b = b
	return nil
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test2")
	if err != nil {
		panic(err)
	}

	err = p.driver.Subscribe("3", "foo")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			msg, err := p.driver.Next()
			if err != nil {
				panic(err)
			}

			if msg == nil {
				continue
			}

			p.log.Info(fmt.Sprintf("%s: %s", Plugin3Name, *msg))
		}
	}()

	return errCh
}

func (p *Plugin3) Stop() error {
	_ = p.driver.Unsubscribe("3", "foo")
	return nil
}

func (p *Plugin3) Name() string {
	return Plugin3Name
}

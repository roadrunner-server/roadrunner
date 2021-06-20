package plugins

import (
	"fmt"

	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const Plugin5Name = "plugin5"

type Plugin5 struct {
	log    logger.Logger
	b      broadcast.Broadcaster
	driver pubsub.SubReader
}

func (p *Plugin5) Init(log logger.Logger, b broadcast.Broadcaster) error {
	p.log = log
	p.b = b
	return nil
}

func (p *Plugin5) Serve() chan error {
	errCh := make(chan error, 1)

	var err error
	p.driver, err = p.b.GetDriver("test4")
	if err != nil {
		panic(err)
	}

	err = p.driver.Subscribe("5", "foo")
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

			p.log.Info(fmt.Sprintf("%s: %s", Plugin5Name, *msg))
		}
	}()

	return errCh
}

func (p *Plugin5) Stop() error {
	_ = p.driver.Unsubscribe("5", "foo")
	return nil
}

func (p *Plugin5) Name() string {
	return Plugin5Name
}

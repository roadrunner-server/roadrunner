package ws

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/broadcast"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	//
	RootPluginName = "broadcast"
	//
	PluginName = "websockets"
)

type Plugin struct {
	// logger
	log logger.Logger
	// configurer plugin
	cfg config.Configurer
}


func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("ws_plugin_init")

	// check for the configuration section existence
	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled, errors.Str("broadcast plugin section should exists in the configuration"))
	}

	p.cfg = cfg
	p.log = log

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error)

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Provides() []interface{} {
	return []interface{}{
		p.Websocket,
	}
}

// Websocket method should provide the Subscriber implementation to the broadcast
func (p *Plugin) Websocket() (broadcast.Subscriber, error) {
	const op = errors.Op("websocket_subscriber_provide")
	ws, err := NewWSSubscriber()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ws, nil
}



func (p *Plugin) Available(){}

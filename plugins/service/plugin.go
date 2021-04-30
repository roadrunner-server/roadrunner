package service

import (
	"sync"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName string = "service"

type Plugin struct {
	sync.Mutex

	logger logger.Logger
	cfg    Config

	// all processes attached to the service
	processes []*Process
}

func (service *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("service_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(errors.Disabled)
	}
	err := cfg.UnmarshalKey(PluginName, &service.cfg.Services)
	if err != nil {
		return errors.E(op, err)
	}

	// init default parameters if not set by user
	service.cfg.InitDefault()
	// save the logger
	service.logger = log

	return nil
}

func (service *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	// start processing
	go func() {
		// lock here, because Stop command might be invoked during the Serve
		service.Lock()
		defer service.Unlock()

		service.processes = make([]*Process, 0, len(service.cfg.Services))
		// for the every service
		for k := range service.cfg.Services {
			// create needed number of the processes
			for i := 0; i < service.cfg.Services[k].ProcessNum; i++ {
				// create processor structure, which will process all the services
				service.processes = append(service.processes, NewServiceProcess(
					service.cfg.Services[k].RemainAfterExit,
					service.cfg.Services[k].ExecTimeout,
					service.cfg.Services[k].RestartSec,
					service.cfg.Services[k].Command,
					service.logger,
					errCh,
				))
			}
		}

		// start all processes
		for i := 0; i < len(service.processes); i++ {
			service.processes[i].start()
		}
	}()

	return errCh
}

func (service *Plugin) Workers() []process.State {
	service.Lock()
	defer service.Unlock()
	states := make([]process.State, 0, len(service.processes))
	for i := 0; i < len(service.processes); i++ {
		st, err := process.GeneralProcessState(service.processes[i].Pid, service.processes[i].rawCmd)
		if err != nil {
			continue
		}
		states = append(states, st)
	}
	return states
}

func (service *Plugin) Stop() error {
	service.Lock()
	defer service.Unlock()

	if len(service.processes) > 0 {
		for i := 0; i < len(service.processes); i++ {
			service.processes[i].stop()
		}
	}
	return nil
}

// Name contains service name.
func (service *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (service *Plugin) Available() {
}

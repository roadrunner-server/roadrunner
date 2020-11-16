package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/interfaces/server"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/util"
)

const ServiceName = "server"

// Plugin manages worker
type Plugin struct {
	cfg     Config
	log     log.Logger
	factory roadrunner.Factory
}

// Init application provider.
func (app *Plugin) Init(cfg config.Configurer, log log.Logger) error {
	const op = errors.Op("Init")
	err := cfg.UnmarshalKey(ServiceName, &app.cfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}
	app.cfg.InitDefaults()
	app.log = log

	return nil
}

// Name contains service name.
func (app *Plugin) Name() string {
	return ServiceName
}

func (app *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	var err error

	app.factory, err = app.initFactory()
	if err != nil {
		errCh <- errors.E(errors.Op("init factory"), err)
	}

	return errCh
}

func (app *Plugin) Stop() error {
	if app.factory == nil {
		return nil
	}

	return app.factory.Close(context.Background())
}

// CmdFactory provides worker command factory assocated with given context.
func (app *Plugin) CmdFactory(env server.Env) (func() *exec.Cmd, error) {
	var cmdArgs []string

	// create command according to the config
	cmdArgs = append(cmdArgs, strings.Split(app.cfg.Command, " ")...)

	return func() *exec.Cmd {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		util.IsolateProcess(cmd)

		// if user is not empty, and OS is linux or macos
		// execute php worker from that particular user
		if app.cfg.User != "" {
			err := util.ExecuteFromUser(cmd, app.cfg.User)
			if err != nil {
				return nil
			}
		}

		cmd.Env = app.setEnv(env)

		return cmd
	}, nil
}

// NewWorker issues new standalone worker.
func (app *Plugin) NewWorker(ctx context.Context, env server.Env) (roadrunner.WorkerBase, error) {
	const op = errors.Op("new worker")
	spawnCmd, err := app.CmdFactory(env)
	if err != nil {
		return nil, errors.E(op, err)
	}

	w, err := app.factory.SpawnWorkerWithContext(ctx, spawnCmd())
	if err != nil {
		return nil, errors.E(op, err)
	}

	w.AddListener(app.collectLogs)

	return w, nil
}

// NewWorkerPool issues new worker pool.
func (app *Plugin) NewWorkerPool(ctx context.Context, opt roadrunner.PoolConfig, env server.Env) (roadrunner.Pool, error) {
	spawnCmd, err := app.CmdFactory(env)
	if err != nil {
		return nil, err
	}

	p, err := roadrunner.NewPool(ctx, spawnCmd, app.factory, opt)
	if err != nil {
		return nil, err
	}

	p.AddListener(app.collectLogs)

	return p, nil
}

// creates relay and worker factory.
func (app *Plugin) initFactory() (roadrunner.Factory, error) {
	const op = errors.Op("network factory init")
	if app.cfg.Relay == "" || app.cfg.Relay == "pipes" {
		return roadrunner.NewPipeFactory(), nil
	}

	dsn := strings.Split(app.cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := util.CreateListener(app.cfg.Relay)
	if err != nil {
		return nil, errors.E(op, errors.Network, err)
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return roadrunner.NewSocketServer(lsn, app.cfg.RelayTimeout), nil
	case "tcp":
		return roadrunner.NewSocketServer(lsn, app.cfg.RelayTimeout), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (app *Plugin) setEnv(e server.Env) []string {
	env := append(os.Environ(), fmt.Sprintf("RR_RELAY=%s", app.cfg.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	return env
}

func (app *Plugin) collectLogs(event interface{}) {
	if we, ok := event.(roadrunner.WorkerEvent); ok {
		switch we.Event {
		case roadrunner.EventWorkerError:
			app.log.Error(we.Payload.(error).Error(), "pid", we.Worker.Pid())
		case roadrunner.EventWorkerLog:
			app.log.Debug(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"), "pid", we.Worker.Pid())
		}
	}
}

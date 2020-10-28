package app

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spiral/endure/errors"
	"github.com/spiral/roadrunner/v2"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/util"
)

const ServiceName = "app"

type Env map[string]string

// WorkerFactory creates workers for the application.
type WorkerFactory interface {
	CmdFactory(env Env) (func() *exec.Cmd, error)
	NewWorker(ctx context.Context, env Env) (roadrunner.WorkerBase, error)
	NewWorkerPool(ctx context.Context, opt roadrunner.Config, env Env) (roadrunner.Pool, error)
}

// App manages worker
type App struct {
	cfg     Config
	log     *zap.Logger
	factory roadrunner.Factory
}

// Init application provider.
func (app *App) Init(cfg config.Provider, log *zap.Logger) error {
	err := cfg.UnmarshalKey(ServiceName, &app.cfg)
	if err != nil {
		return err
	}
	app.cfg.InitDefaults()
	app.log = log

	return nil
}

// Name contains service name.
func (app *App) Name() string {
	return ServiceName
}

func (app *App) Serve() chan error {
	errCh := make(chan error, 1)
	var err error

	app.factory, err = app.initFactory()
	if err != nil {
		errCh <- errors.E(errors.Op("init factory"), err)
	}

	app.log.Debug("Started worker factory", zap.Any("relay", app.cfg.Relay), zap.Any("command", app.cfg.Command))

	return errCh
}

func (app *App) Stop() error {
	if app.factory == nil {
		return nil
	}

	return app.factory.Close(context.Background())
}

// CmdFactory provides worker command factory assocated with given context.
func (app *App) CmdFactory(env Env) (func() *exec.Cmd, error) {
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
func (app *App) NewWorker(ctx context.Context, env Env) (roadrunner.WorkerBase, error) {
	spawnCmd, err := app.CmdFactory(env)
	if err != nil {
		return nil, err
	}

	return app.factory.SpawnWorkerWithContext(ctx, spawnCmd())
}

// NewWorkerPool issues new worker pool.
func (app *App) NewWorkerPool(ctx context.Context, opt roadrunner.Config, env Env) (roadrunner.Pool, error) {
	spawnCmd, err := app.CmdFactory(env)
	if err != nil {
		return nil, err
	}

	p, err := roadrunner.NewPool(ctx, spawnCmd, app.factory, opt)
	if err != nil {
		return nil, err
	}

	p.AddListener(func(event interface{}) {
		if we, ok := event.(roadrunner.WorkerEvent); ok {
			if we.Event == roadrunner.EventWorkerLog {
				log.Print(color.YellowString(string(we.Payload.([]byte))))
			}
		}
	})

	return p, nil
}

// creates relay and worker factory.
func (app *App) initFactory() (roadrunner.Factory, error) {
	if app.cfg.Relay == "" || app.cfg.Relay == "pipes" {
		return roadrunner.NewPipeFactory(), nil
	}

	dsn := strings.Split(app.cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := util.CreateListener(app.cfg.Relay)
	if err != nil {
		return nil, err
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return roadrunner.NewSocketServer(lsn, app.cfg.RelayTimeout), nil
	case "tcp":
		return roadrunner.NewSocketServer(lsn, app.cfg.RelayTimeout), nil
	default:
		return nil, errors.E(errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (app *App) setEnv(e Env) []string {
	env := append(os.Environ(), fmt.Sprintf("RR_RELAY=%s", app.cfg.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	return env
}

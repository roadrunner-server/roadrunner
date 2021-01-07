package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"

	// core imports
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/pool"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	"github.com/spiral/roadrunner/v2/pkg/pipe"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/socket"
	"github.com/spiral/roadrunner/v2/utils"
)

// PluginName for the server
const PluginName = "server"

// Plugin manages worker
type Plugin struct {
	cfg     Config
	log     logger.Logger
	factory worker.Factory
}

// Init application provider.
func (server *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("Init")
	err := cfg.UnmarshalKey(PluginName, &server.cfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}
	server.cfg.InitDefaults()
	server.log = log

	server.factory, err = server.initFactory()
	if err != nil {
		return errors.E(errors.Op("Init factory"), err)
	}

	return nil
}

// Name contains service name.
func (server *Plugin) Name() string {
	return PluginName
}

// Serve (Start) server plugin (just a mock here to satisfy interface)
func (server *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

// Stop used to close chosen in config factory
func (server *Plugin) Stop() error {
	if server.factory == nil {
		return nil
	}

	return server.factory.Close()
}

// CmdFactory provides worker command factory associated with given context.
func (server *Plugin) CmdFactory(env Env) (func() *exec.Cmd, error) {
	const op = errors.Op("cmd factory")
	var cmdArgs []string

	// create command according to the config
	cmdArgs = append(cmdArgs, strings.Split(server.cfg.Command, " ")...)
	if len(cmdArgs) < 2 {
		return nil, errors.E(op, errors.Str("should be in form of `php <script>"))
	}
	if cmdArgs[0] != "php" {
		return nil, errors.E(op, errors.Str("first arg in command should be `php`"))
	}

	_, err := os.Stat(cmdArgs[1])
	if err != nil {
		return nil, errors.E(op, err)
	}
	return func() *exec.Cmd {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
		utils.IsolateProcess(cmd)

		// if user is not empty, and OS is linux or macos
		// execute php worker from that particular user
		if server.cfg.User != "" {
			err := utils.ExecuteFromUser(cmd, server.cfg.User)
			if err != nil {
				return nil
			}
		}

		cmd.Env = server.setEnv(env)

		return cmd
	}, nil
}

// NewWorker issues new standalone worker.
func (server *Plugin) NewWorker(ctx context.Context, env Env, listeners ...events.Listener) (worker.BaseProcess, error) {
	const op = errors.Op("new worker")

	list := make([]events.Listener, 0, len(listeners))
	list = append(list, server.collectWorkerLogs)

	spawnCmd, err := server.CmdFactory(env)
	if err != nil {
		return nil, errors.E(op, err)
	}

	w, err := server.factory.SpawnWorkerWithTimeout(ctx, spawnCmd(), list...)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return w, nil
}

// NewWorkerPool issues new worker pool.
func (server *Plugin) NewWorkerPool(ctx context.Context, opt poolImpl.Config, env Env, listeners ...events.Listener) (pool.Pool, error) {
	const op = errors.Op("server plugins new worker pool")
	spawnCmd, err := server.CmdFactory(env)
	if err != nil {
		return nil, errors.E(op, err)
	}

	list := make([]events.Listener, 0, 1)
	list = append(list, server.collectPoolLogs)
	if len(listeners) != 0 {
		list = append(list, listeners...)
	}

	p, err := poolImpl.Initialize(ctx, spawnCmd, server.factory, opt, poolImpl.AddListeners(list...))
	if err != nil {
		return nil, errors.E(op, err)
	}

	return p, nil
}

// creates relay and worker factory.
func (server *Plugin) initFactory() (worker.Factory, error) {
	const op = errors.Op("network factory init")
	if server.cfg.Relay == "" || server.cfg.Relay == "pipes" {
		return pipe.NewPipeFactory(), nil
	}

	dsn := strings.Split(server.cfg.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := utils.CreateListener(server.cfg.Relay)
	if err != nil {
		return nil, errors.E(op, errors.Network, err)
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return socket.NewSocketServer(lsn, server.cfg.RelayTimeout), nil
	case "tcp":
		return socket.NewSocketServer(lsn, server.cfg.RelayTimeout), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (server *Plugin) setEnv(e Env) []string {
	env := append(os.Environ(), fmt.Sprintf("RR_RELAY=%s", server.cfg.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	return env
}

func (server *Plugin) collectPoolLogs(event interface{}) {
	if we, ok := event.(events.PoolEvent); ok {
		switch we.Event {
		case events.EventMaxMemory:
			server.log.Info("worker max memory reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventNoFreeWorkers:
			server.log.Info("no free workers in pool", "error", we.Payload.(error).Error())
		case events.EventPoolError:
			server.log.Info("pool error", "error", we.Payload.(error).Error())
		case events.EventSupervisorError:
			server.log.Info("pool supervizor error", "error", we.Payload.(error).Error())
		case events.EventTTL:
			server.log.Info("worker TTL reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventWorkerConstruct:
			if _, ok := we.Payload.(error); ok {
				server.log.Error("worker construction error", "error", we.Payload.(error).Error())
				return
			}
			server.log.Info("worker constructed", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventWorkerDestruct:
			server.log.Info("worker destructed", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventExecTTL:
			server.log.Info("EVENT EXEC TTL PLACEHOLDER")
		case events.EventIdleTTL:
			server.log.Info("worker IDLE timeout reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		}
	}

	if we, ok := event.(events.WorkerEvent); ok {
		switch we.Event {
		case events.EventWorkerError:
			server.log.Info(we.Payload.(error).Error(), "pid", we.Worker.(worker.BaseProcess).Pid())
		case events.EventWorkerLog:
			server.log.Info(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"), "pid", we.Worker.(worker.BaseProcess).Pid())
		}
	}
}

func (server *Plugin) collectWorkerLogs(event interface{}) {
	if we, ok := event.(events.WorkerEvent); ok {
		switch we.Event {
		case events.EventWorkerError:
			server.log.Error(we.Payload.(error).Error(), "pid", we.Worker.(worker.BaseProcess).Pid())
		case events.EventWorkerLog:
			server.log.Info(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"), "pid", we.Worker.(worker.BaseProcess).Pid())
		}
	}
}

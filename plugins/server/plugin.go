package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/transport"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"

	// core imports
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	"github.com/spiral/roadrunner/v2/pkg/transport/socket"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/utils"
)

// PluginName for the server
const PluginName = "server"

// RR_RELAY env variable key (internal)
const RR_RELAY = "RR_RELAY" //nolint:golint,stylecheck
// RR_RPC env variable key (internal) if the RPC presents
const RR_RPC = "RR_RPC" //nolint:golint,stylecheck

// Plugin manages worker
type Plugin struct {
	cfg     Config
	log     logger.Logger
	factory transport.Factory
}

// Init application provider.
func (server *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("server_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}
	err := cfg.Unmarshal(&server.cfg)
	if err != nil {
		return errors.E(op, errors.Init, err)
	}
	server.cfg.InitDefaults()
	server.log = log

	return nil
}

// Name contains service name.
func (server *Plugin) Name() string {
	return PluginName
}

// Serve (Start) server plugin (just a mock here to satisfy interface)
func (server *Plugin) Serve() chan error {
	const op = errors.Op("server_plugin_serve")
	errCh := make(chan error, 1)
	var err error
	server.factory, err = server.initFactory()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}
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
	const op = errors.Op("server_plugin_cmd_factory")
	var cmdArgs []string

	// create command according to the config
	cmdArgs = append(cmdArgs, strings.Split(server.cfg.Server.Command, " ")...)
	if len(cmdArgs) < 2 {
		return nil, errors.E(op, errors.Str("minimum command should be `<executable> <script>"))
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
		if server.cfg.Server.User != "" {
			err := utils.ExecuteFromUser(cmd, server.cfg.Server.User)
			if err != nil {
				return nil
			}
		}

		cmd.Env = server.setEnv(env)

		return cmd
	}, nil
}

// NewWorker issues new standalone worker.
func (server *Plugin) NewWorker(ctx context.Context, env Env, listeners ...events.Listener) (*worker.Process, error) {
	const op = errors.Op("server_plugin_new_worker")

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
func (server *Plugin) NewWorkerPool(ctx context.Context, opt pool.Config, env Env, listeners ...events.Listener) (pool.Pool, error) {
	const op = errors.Op("server_plugin_new_worker_pool")
	spawnCmd, err := server.CmdFactory(env)
	if err != nil {
		return nil, errors.E(op, err)
	}

	list := make([]events.Listener, 0, 1)
	list = append(list, server.collectEvents)
	if len(listeners) != 0 {
		list = append(list, listeners...)
	}

	p, err := pool.Initialize(ctx, spawnCmd, server.factory, opt, pool.AddListeners(list...))
	if err != nil {
		return nil, errors.E(op, err)
	}

	return p, nil
}

// creates relay and worker factory.
func (server *Plugin) initFactory() (transport.Factory, error) {
	const op = errors.Op("server_plugin_init_factory")
	if server.cfg.Server.Relay == "" || server.cfg.Server.Relay == "pipes" {
		return pipe.NewPipeFactory(), nil
	}

	dsn := strings.Split(server.cfg.Server.Relay, "://")
	if len(dsn) != 2 {
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}

	lsn, err := utils.CreateListener(server.cfg.Server.Relay)
	if err != nil {
		return nil, errors.E(op, errors.Network, err)
	}

	switch dsn[0] {
	// sockets group
	case "unix":
		return socket.NewSocketServer(lsn, server.cfg.Server.RelayTimeout), nil
	case "tcp":
		return socket.NewSocketServer(lsn, server.cfg.Server.RelayTimeout), nil
	default:
		return nil, errors.E(op, errors.Network, errors.Str("invalid DSN (tcp://:6001, unix://file.sock)"))
	}
}

func (server *Plugin) setEnv(e Env) []string {
	env := append(os.Environ(), fmt.Sprintf(RR_RELAY+"=%s", server.cfg.Server.Relay))
	for k, v := range e {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}

	if server.cfg.RPC != nil && server.cfg.RPC.Listen != "" {
		env = append(env, fmt.Sprintf("%s=%s", RR_RPC, server.cfg.RPC.Listen))
	}

	// set env variables from the config
	if len(server.cfg.Server.Env) > 0 {
		for k, v := range server.cfg.Server.Env {
			env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
	}

	return env
}

func (server *Plugin) collectEvents(event interface{}) {
	if we, ok := event.(events.PoolEvent); ok {
		switch we.Event {
		case events.EventMaxMemory:
			server.log.Warn("worker max memory reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventNoFreeWorkers:
			server.log.Warn("no free workers in pool", "error", we.Payload.(error).Error())
		case events.EventPoolError:
			server.log.Error("pool error", "error", we.Payload.(error).Error())
		case events.EventSupervisorError:
			server.log.Error("pool supervisor error", "error", we.Payload.(error).Error())
		case events.EventTTL:
			server.log.Warn("worker TTL reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventWorkerConstruct:
			if _, ok := we.Payload.(error); ok {
				server.log.Error("worker construction error", "error", we.Payload.(error).Error())
				return
			}
			server.log.Debug("worker constructed", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventWorkerDestruct:
			server.log.Debug("worker destructed", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventExecTTL:
			server.log.Warn("worker exec timeout reached", "error", we.Payload.(error).Error())
		case events.EventIdleTTL:
			server.log.Warn("worker idle timeout reached", "pid", we.Payload.(worker.BaseProcess).Pid())
		case events.EventPoolRestart:
			server.log.Warn("requested pool restart")
		}
	}

	if we, ok := event.(events.WorkerEvent); ok {
		switch we.Event {
		case events.EventWorkerError:
			server.log.Error(strings.TrimRight(we.Payload.(error).Error(), " \n\t"))
		case events.EventWorkerLog:
			server.log.Debug(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"))
		case events.EventWorkerStderr:
			// TODO unsafe byte to string convert
			server.log.Debug(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"))
		}
	}
}

func (server *Plugin) collectWorkerLogs(event interface{}) {
	if we, ok := event.(events.WorkerEvent); ok {
		switch we.Event {
		case events.EventWorkerError:
			server.log.Error(strings.TrimRight(we.Payload.(error).Error(), " \n\t"))
		case events.EventWorkerLog:
			server.log.Debug(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"))
		case events.EventWorkerStderr:
			// TODO unsafe byte to string convert
			server.log.Debug(strings.TrimRight(string(we.Payload.([]byte)), " \n\t"))
		}
	}
}

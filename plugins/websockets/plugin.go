package websockets

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	phpPool "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/process"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/pkg/pubsub/message"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/executor"
	"github.com/spiral/roadrunner/v2/plugins/websockets/pool"
	"github.com/spiral/roadrunner/v2/plugins/websockets/storage"
	"github.com/spiral/roadrunner/v2/plugins/websockets/validator"
	"github.com/spiral/roadrunner/v2/utils"
)

const (
	PluginName string = "websockets"
)

type Plugin struct {
	sync.RWMutex
	// Collection with all available pubsubs
	pubsubs map[string]pubsub.PubSub

	cfg *Config
	log logger.Logger

	// global connections map
	connections sync.Map
	storage     *storage.Storage

	// GO workers pool
	workersPool *pool.WorkersPool

	wsUpgrade *websocket.Upgrader
	serveExit chan struct{}

	phpPool phpPool.Pool
	server  server.Server

	// function used to validate access to the requested resource
	accessValidator validator.AccessValidatorFn
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger, server server.Server) error {
	const op = errors.Op("websockets_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	p.cfg.InitDefault()

	p.pubsubs = make(map[string]pubsub.PubSub)
	p.log = log
	p.storage = storage.NewStorage()
	p.workersPool = pool.NewWorkersPool(p.storage, &p.connections, log)
	p.wsUpgrade = &websocket.Upgrader{
		HandshakeTimeout: time.Second * 60,
	}
	p.serveExit = make(chan struct{})
	p.server = server

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error)

	go func() {
		var err error
		p.Lock()
		defer p.Unlock()

		p.phpPool, err = p.server.NewWorkerPool(context.Background(), phpPool.Config{
			Debug:           p.cfg.Pool.Debug,
			NumWorkers:      p.cfg.Pool.NumWorkers,
			MaxJobs:         p.cfg.Pool.MaxJobs,
			AllocateTimeout: p.cfg.Pool.AllocateTimeout,
			DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
			Supervisor:      p.cfg.Pool.Supervisor,
		}, map[string]string{"RR_MODE": "http"})
		if err != nil {
			errCh <- err
		}

		p.accessValidator = p.defaultAccessValidator(p.phpPool)
	}()

	// run all pubsubs drivers
	for _, v := range p.pubsubs {
		go func(ps pubsub.PubSub) {
			for {
				select {
				case <-p.serveExit:
					return
				default:
					data, err := ps.Next()
					if err != nil {
						errCh <- err
						return
					}
					p.workersPool.Queue(data)
				}
			}
		}(v)
	}

	return errCh
}

func (p *Plugin) Stop() error {
	// close workers pool
	p.workersPool.Stop()
	p.Lock()
	p.phpPool.Destroy(context.Background())
	p.Unlock()
	return nil
}

func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.GetPublishers,
	}
}

func (p *Plugin) Available() {}

func (p *Plugin) RPC() interface{} {
	return &rpc{
		plugin: p,
		log:    p.log,
	}
}

func (p *Plugin) Name() string {
	return PluginName
}

// GetPublishers collects all pubsubs
func (p *Plugin) GetPublishers(name endure.Named, pub pubsub.PubSub) {
	p.pubsubs[name.Name()] = pub
}

func (p *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != p.cfg.Path {
			next.ServeHTTP(w, r)
			return
		}

		// we need to lock here, because accessValidator might not be set in the Serve func at the moment
		p.RLock()
		// before we hijacked connection, we still can write to the response headers
		val, err := p.accessValidator(r)
		p.RUnlock()
		if err != nil {
			p.log.Error("validation error")
			w.WriteHeader(400)
			return
		}

		if val.Status != http.StatusOK {
			for k, v := range val.Header {
				for i := 0; i < len(v); i++ {
					w.Header().Add(k, v[i])
				}
			}
			w.WriteHeader(val.Status)
			_, _ = w.Write(val.Body)
			return
		}

		// upgrade connection to websocket connection
		_conn, err := p.wsUpgrade.Upgrade(w, r, nil)
		if err != nil {
			// connection hijacked, do not use response.writer or request
			p.log.Error("upgrade connection", "error", err)
			return
		}

		// construct safe connection protected by mutexes
		safeConn := connection.NewConnection(_conn, p.log)
		// generate UUID from the connection
		connectionID := uuid.NewString()
		// store connection
		p.connections.Store(connectionID, safeConn)

		defer func() {
			// close the connection on exit
			err = safeConn.Close()
			if err != nil {
				p.log.Error("connection close", "error", err)
			}

			// when exiting - delete the connection
			p.connections.Delete(connectionID)
		}()

		// Executor wraps a connection to have a safe abstraction
		e := executor.NewExecutor(safeConn, p.log, p.storage, connectionID, p.pubsubs, p.accessValidator, r)
		p.log.Info("websocket client connected", "uuid", connectionID)
		defer e.CleanUp()

		err = e.StartCommandLoop()
		if err != nil {
			p.log.Error("command loop error, disconnecting", "error", err.Error())
			return
		}

		p.log.Info("disconnected", "connectionID", connectionID)
	})
}

// Workers returns slice with the process states for the workers
func (p *Plugin) Workers() []process.State {
	p.RLock()
	defer p.RUnlock()

	workers := p.workers()

	ps := make([]process.State, 0, len(workers))
	for i := 0; i < len(workers); i++ {
		state, err := process.WorkerProcessState(workers[i])
		if err != nil {
			return nil
		}
		ps = append(ps, state)
	}

	return ps
}

// internal
func (p *Plugin) workers() []worker.BaseProcess {
	return p.phpPool.Workers()
}

// Reset destroys the old pool and replaces it with new one, waiting for old pool to die
func (p *Plugin) Reset() error {
	p.Lock()
	defer p.Unlock()
	const op = errors.Op("ws_plugin_reset")
	p.log.Info("WS plugin got restart request. Restarting...")
	p.phpPool.Destroy(context.Background())
	p.phpPool = nil

	var err error
	p.phpPool, err = p.server.NewWorkerPool(context.Background(), phpPool.Config{
		Debug:           p.cfg.Pool.Debug,
		NumWorkers:      p.cfg.Pool.NumWorkers,
		MaxJobs:         p.cfg.Pool.MaxJobs,
		AllocateTimeout: p.cfg.Pool.AllocateTimeout,
		DestroyTimeout:  p.cfg.Pool.DestroyTimeout,
		Supervisor:      p.cfg.Pool.Supervisor,
	}, map[string]string{"RR_MODE": "http"})
	if err != nil {
		return errors.E(op, err)
	}

	// attach validators
	p.accessValidator = p.defaultAccessValidator(p.phpPool)

	p.log.Info("WS plugin successfully restarted")
	return nil
}

// Publish is an entry point to the websocket PUBSUB
func (p *Plugin) Publish(m []byte) error {
	p.Lock()
	defer p.Unlock()

	// Get payload
	fbsMsg := message.GetRootAsMessages(m, 0)
	tmpMsg := &message.Message{}

	for i := 0; i < fbsMsg.MessagesLength(); i++ {
		fbsMsg.Messages(tmpMsg, i)

		for j := 0; j < tmpMsg.TopicsLength(); j++ {
			if br, ok := p.pubsubs[utils.AsString(tmpMsg.Broker())]; ok {
				table := tmpMsg.Table()
				err := br.Publish(table.ByteVector(0))
				if err != nil {
					return errors.E(err)
				}
			} else {
				p.log.Warn("no such broker", "available", p.pubsubs, "requested", tmpMsg.Broker())
			}
		}
	}
	return nil
}

func (p *Plugin) PublishAsync(msg []byte) {
	//go func() {
	//	p.Lock()
	//	defer p.Unlock()
	//	for i := 0; i < len(msg); i++ {
	//		for j := 0; j < len(msg[i].Topics); j++ {
	//			err := p.pubsubs[msg[i].Broker].Publish(msg)
	//			if err != nil {
	//				p.log.Error("publish async error", "error", err)
	//				return
	//			}
	//		}
	//	}
	//}()
}

func (p *Plugin) defaultAccessValidator(pool phpPool.Pool) validator.AccessValidatorFn {
	return func(r *http.Request, topics ...string) (*validator.AccessValidator, error) {
		p.RLock()
		defer p.RUnlock()
		const op = errors.Op("access_validator")

		p.log.Debug("validation", "topics", topics)
		r = attributes.Init(r)

		// if channels len is eq to 0, we use serverValidator
		if len(topics) == 0 {
			ctx, err := validator.ServerAccessValidator(r)
			if err != nil {
				return nil, errors.E(op, err)
			}

			val, err := exec(ctx, pool)
			if err != nil {
				return nil, errors.E(err)
			}

			return val, nil
		}

		ctx, err := validator.TopicsAccessValidator(r, topics...)
		if err != nil {
			return nil, errors.E(op, err)
		}

		val, err := exec(ctx, pool)
		if err != nil {
			return nil, errors.E(op)
		}

		if val.Status != http.StatusOK {
			return val, errors.E(op, errors.Errorf("access forbidden, code: %d", val.Status))
		}

		return val, nil
	}
}

// go:inline
func exec(ctx []byte, pool phpPool.Pool) (*validator.AccessValidator, error) {
	const op = errors.Op("exec")
	pd := payload.Payload{
		Context: ctx,
	}

	resp, err := pool.Exec(pd)
	if err != nil {
		return nil, errors.E(op, err)
	}

	val := &validator.AccessValidator{
		Body: resp.Body,
	}

	err = json.Unmarshal(resp.Context, val)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return val, nil
}

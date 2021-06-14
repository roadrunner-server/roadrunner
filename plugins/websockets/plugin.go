package websockets

import (
	"context"
	"fmt"
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
	websocketsv1 "github.com/spiral/roadrunner/v2/pkg/proto/websockets/v1beta"
	"github.com/spiral/roadrunner/v2/pkg/pubsub"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/spiral/roadrunner/v2/plugins/websockets/connection"
	"github.com/spiral/roadrunner/v2/plugins/websockets/executor"
	"github.com/spiral/roadrunner/v2/plugins/websockets/pool"
	"github.com/spiral/roadrunner/v2/plugins/websockets/validator"
	"google.golang.org/protobuf/proto"
)

const (
	PluginName string = "websockets"
)

type Plugin struct {
	sync.RWMutex
	// Collection with all available pubsubs
	pubsubs map[string]pubsub.PubSub

	psProviders map[string]pubsub.PSProvider

	cfg       *Config
	cfgPlugin config.Configurer
	log       logger.Logger

	// global connections map
	connections sync.Map

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
	p.psProviders = make(map[string]pubsub.PSProvider)

	p.log = log
	p.cfgPlugin = cfg

	p.wsUpgrade = &websocket.Upgrader{
		HandshakeTimeout: time.Second * 60,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}
	p.serveExit = make(chan struct{})
	p.server = server

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error)
	const op = errors.Op("websockets_plugin_serve")

	err := p.initPubSubs()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

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

	p.workersPool = pool.NewWorkersPool(p.pubsubs, &p.connections, p.log)

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

func (p *Plugin) initPubSubs() error {
	for i := 0; i < len(p.cfg.PubSubs); i++ {
		// don't need to have a section for the in-memory
		if p.cfg.PubSubs[i] == "memory" {
			if provider, ok := p.psProviders[p.cfg.PubSubs[i]]; ok {
				r, err := provider.PSProvide("")
				if err != nil {
					return err
				}

				// append default in-memory provider
				p.pubsubs["memory"] = r
			}
			continue
		}
		// key - memory, redis
		if provider, ok := p.psProviders[p.cfg.PubSubs[i]]; ok {
			// try local key
			switch {
			// try local config first
			case p.cfgPlugin.Has(fmt.Sprintf("%s.%s", PluginName, p.cfg.PubSubs[i])):
				r, err := provider.PSProvide(fmt.Sprintf("%s.%s", PluginName, p.cfg.PubSubs[i]))
				if err != nil {
					return err
				}

				// append redis provider
				p.pubsubs[p.cfg.PubSubs[i]] = r
			case p.cfgPlugin.Has(p.cfg.PubSubs[i]):
				r, err := provider.PSProvide(p.cfg.PubSubs[i])
				if err != nil {
					return err
				}

				// append redis provider
				p.pubsubs[p.cfg.PubSubs[i]] = r
			default:
				return errors.Errorf("could not find configuration sections for the %s", p.cfg.PubSubs[i])
			}
		} else {
			// no such driver
			p.log.Warn("no such driver", "requested", p.cfg.PubSubs[i], "available", p.psProviders)
		}
	}

	return nil
}

func (p *Plugin) Stop() error {
	// close workers pool
	p.workersPool.Stop()
	p.Lock()
	if p.phpPool == nil {
		p.Unlock()
		return nil
	}
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
func (p *Plugin) GetPublishers(name endure.Named, pub pubsub.PSProvider) {
	p.psProviders[name.Name()] = pub
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

		// Executor wraps a connection to have a safe abstraction
		e := executor.NewExecutor(safeConn, p.log, connectionID, p.pubsubs, p.accessValidator, r)
		p.log.Info("websocket client connected", "uuid", connectionID)

		err = e.StartCommandLoop()
		if err != nil {
			p.log.Error("command loop error, disconnecting", "error", err.Error())
			return
		}

		// when exiting - delete the connection
		p.connections.Delete(connectionID)

		// remove connection from all topics from all pub-sub drivers
		e.CleanUp()

		err = r.Body.Close()
		if err != nil {
			p.log.Error("body close", "error", err)
		}

		// close the connection on exit
		err = safeConn.Close()
		if err != nil {
			p.log.Error("connection close", "error", err)
		}

		safeConn = nil
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

	msg := &websocketsv1.Message{}
	err := proto.Unmarshal(m, msg)
	if err != nil {
		return err
	}

	// Get payload
	for i := 0; i < len(msg.GetTopics()); i++ {
		if br, ok := p.pubsubs[msg.GetBroker()]; ok {
			err := br.Publish(m)
			if err != nil {
				return errors.E(err)
			}
		} else {
			p.log.Warn("no such broker", "available", p.pubsubs, "requested", msg.GetBroker())
		}
	}
	return nil
}

func (p *Plugin) PublishAsync(m []byte) {
	go func() {
		p.Lock()
		defer p.Unlock()
		msg := &websocketsv1.Message{}
		err := proto.Unmarshal(m, msg)
		if err != nil {
			p.log.Error("message unmarshal")
		}

		// Get payload
		for i := 0; i < len(msg.GetTopics()); i++ {
			if br, ok := p.pubsubs[msg.GetBroker()]; ok {
				err := br.Publish(m)
				if err != nil {
					p.log.Error("publish async error", "error", err)
				}
			} else {
				p.log.Warn("no such broker", "available", p.pubsubs, "requested", msg.GetBroker())
			}
		}
	}()
}

func (p *Plugin) defaultAccessValidator(pool phpPool.Pool) validator.AccessValidatorFn {
	return func(r *http.Request, topics ...string) (*validator.AccessValidator, error) {
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

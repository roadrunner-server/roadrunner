package websockets

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// ID defines service id.
const ID = "ws"

// Service to manage websocket clients.
type Service struct {
	cfg       *Config
	upgrade   websocket.Upgrader
	client    *broadcast.Client
	connPool  *connPool
	listeners []func(event int, ctx interface{})
	mu        sync.Mutex
	stopped   int32
	stop      chan error
}

// AddListener attaches server event controller.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.listeners = append(s.listeners, l)
}

// Init the service.
func (s *Service) Init(
	cfg *Config,
	env env.Environment,
	rttp *rhttp.Service,
	rpc *rpc.Service,
	broadcast *broadcast.Service,
) (bool, error) {
	if broadcast == nil || rpc == nil {
		// unable to activate
		return false, nil
	}

	s.cfg = cfg
	s.client = broadcast.NewClient()
	s.connPool = newPool(s.client, s.reportError)
	s.stopped = 0

	if err := rpc.Register(ID, &rpcService{svc: s}); err != nil {
		return false, err
	}

	if env != nil {
		// ensure that underlying kernel knows what route to handle
		env.SetEnv("RR_BROADCAST_PATH", cfg.Path)
	}

	// init all this stuff
	s.upgrade = websocket.Upgrader{}

	if s.cfg.NoOrigin {
		s.upgrade.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	rttp.AddMiddleware(s.middleware)

	return true, nil
}

// Serve the websocket connections.
func (s *Service) Serve() error {
	defer s.client.Close()
	defer s.connPool.close()

	s.mu.Lock()
	s.stop = make(chan error)
	s.mu.Unlock()

	return <-s.stop
}

// Stop the service and disconnect all connections.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		close(s.stop)
	}
}

// middleware intercepts websocket connections.
func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != s.cfg.Path {
			f(w, r)
			return
		}

		// checking server access
		if err := newValidator().assertServerAccess(f, r); err != nil {
			// show the error to the user
			if av, ok := err.(*accessValidator); ok {
				av.copy(w)
			} else {
				w.WriteHeader(400)
			}
			return
		}

		conn, err := s.upgrade.Upgrade(w, r, nil)
		if err != nil {
			s.reportError(err, nil)
			return
		}

		s.throw(EventConnect, conn)

		// manage connection
		ctx, err := s.connPool.connect(conn)
		if err != nil {
			s.reportError(err, conn)
			return
		}

		s.serveConn(ctx, f, r)
	}
}

// send and receive messages over websocket
func (s *Service) serveConn(ctx *ConnContext, f http.HandlerFunc, r *http.Request) {
	defer func() {
		if err := s.connPool.disconnect(ctx.Conn); err != nil {
			s.reportError(err, ctx.Conn)
		}
		s.throw(EventDisconnect, ctx.Conn)
	}()

	s.handleCommands(ctx, f, r)
}

func (s *Service) handleCommands(ctx *ConnContext, f http.HandlerFunc, r *http.Request) {
	cmd := &broadcast.Message{}
	for {
		if err := ctx.Conn.ReadJSON(cmd); err != nil {
			s.reportError(err, ctx.Conn)
			return
		}

		switch cmd.Topic {
		case "join":
			topics := make([]string, 0)
			if err := unmarshalCommand(cmd, &topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if len(topics) == 0 {
				continue
			}

			if err := newValidator().assertTopicsAccess(f, r, topics...); err != nil {
				s.reportError(err, ctx.Conn)

				if err := ctx.SendMessage("#join", topics); err != nil {
					s.reportError(err, ctx.Conn)
					return
				}

				continue
			}

			if err := s.connPool.subscribe(ctx, topics...); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if err := ctx.SendMessage("@join", topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			s.throw(EventJoin, &TopicEvent{Conn: ctx.Conn, Topics: topics})
		case "leave":
			topics := make([]string, 0)
			if err := unmarshalCommand(cmd, &topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if len(topics) == 0 {
				continue
			}

			if err := s.connPool.unsubscribe(ctx, topics...); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if err := ctx.SendMessage("@leave", topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			s.throw(EventLeave, &TopicEvent{Conn: ctx.Conn, Topics: topics})
		}
	}
}

// handle connection error
func (s *Service) reportError(err error, conn *websocket.Conn) {
	s.throw(EventError, &ErrorEvent{Conn: conn, Error: err})
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.listeners {
		l(event, ctx)
	}
}

// unmarshalCommand command data.
func unmarshalCommand(msg *broadcast.Message, v interface{}) error {
	return json.Unmarshal(msg.Payload, v)
}

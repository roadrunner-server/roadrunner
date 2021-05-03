package websockets

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/spiral/broadcast/v2"
)

// manages a set of websocket connections
type connPool struct {
	errHandler func(err error, conn *websocket.Conn)

	mur    sync.Mutex
	client *broadcast.Client
	router *broadcast.Router

	mu    sync.Mutex
	conns map[*websocket.Conn]*ConnContext
}

// create new connection pool
func newPool(client *broadcast.Client, errHandler func(err error, conn *websocket.Conn)) *connPool {
	cp := &connPool{
		client:     client,
		router:     broadcast.NewRouter(),
		errHandler: errHandler,
		conns:      map[*websocket.Conn]*ConnContext{},
	}

	go func() {
		for msg := range cp.client.Channel() {
			cp.mur.Lock()
			cp.router.Dispatch(msg)
			cp.mur.Unlock()
		}
	}()

	return cp
}

// connect the websocket and register client in message router
func (cp *connPool) connect(conn *websocket.Conn) (*ConnContext, error) {
	ctx := &ConnContext{
		Conn:     conn,
		Topics:   []string{},
		upstream: make(chan *broadcast.Message),
	}

	cp.mu.Lock()
	cp.conns[conn] = ctx
	cp.mu.Unlock()

	go ctx.serve(cp.errHandler)

	return ctx, nil
}

// disconnect the websocket
func (cp *connPool) disconnect(conn *websocket.Conn) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	ctx, ok := cp.conns[conn]
	if !ok {
		return errors.New("no such connection")
	}

	if err := cp.unsubscribe(ctx, ctx.Topics...); err != nil {
		cp.errHandler(err, conn)
	}

	delete(cp.conns, conn)

	return conn.Close()
}

// subscribe the connection
func (cp *connPool) subscribe(ctx *ConnContext, topics ...string) error {
	cp.mur.Lock()
	defer cp.mur.Unlock()

	ctx.addTopics(topics...)

	newTopics := cp.router.Subscribe(ctx.upstream, topics...)
	if len(newTopics) != 0 {
		return cp.client.Subscribe(newTopics...)
	}

	return nil
}

// unsubscribe the connection
func (cp *connPool) unsubscribe(ctx *ConnContext, topics ...string) error {
	cp.mur.Lock()
	defer cp.mur.Unlock()

	ctx.dropTopic(topics...)

	dropTopics := cp.router.Unsubscribe(ctx.upstream, topics...)
	if len(dropTopics) != 0 {
		return cp.client.Unsubscribe(dropTopics...)
	}

	return nil
}

// close the connection pool and disconnect all listeners
func (cp *connPool) close() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	for conn, ctx := range cp.conns {
		if err := cp.unsubscribe(ctx, ctx.Topics...); err != nil {
			cp.errHandler(err, conn)
		}

		delete(cp.conns, conn)

		if err := conn.Close(); err != nil {
			cp.errHandler(err, conn)
		}
	}
}

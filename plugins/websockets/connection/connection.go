package connection

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// Connection represents wrapped and safe to use from the different threads websocket connection
type Connection struct {
	sync.RWMutex
	log  logger.Logger
	conn *websocket.Conn
}

func NewConnection(wsConn *websocket.Conn, log logger.Logger) *Connection {
	return &Connection{
		conn: wsConn,
		log:  log,
	}
}

func (c *Connection) Write(data []byte) error {
	c.Lock()
	defer c.Unlock()

	const op = errors.Op("websocket_write")
	// handle a case when a goroutine tried to write into the closed connection
	defer func() {
		if r := recover(); r != nil {
			c.log.Warn("panic handled, tried to write into the closed connection")
		}
	}()

	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *Connection) Read() (int, []byte, error) {
	const op = errors.Op("websocket_read")

	mt, data, err := c.conn.ReadMessage()
	if err != nil {
		return -1, nil, errors.E(op, err)
	}

	return mt, data, nil
}

func (c *Connection) Close() error {
	c.Lock()
	defer c.Unlock()
	const op = errors.Op("websocket_close")

	err := c.conn.Close()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

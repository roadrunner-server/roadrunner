package amqp

import (
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

// manages set of AMQP channels
type chanPool struct {
	// timeout to backoff redial
	tout time.Duration
	url  string

	mu *sync.Mutex

	conn      *amqp.Connection
	channels  map[string]*channel
	wait      chan interface{}
	connected chan interface{}
}

// manages single channel
type channel struct {
	ch *amqp.Channel
	// todo unused
	//consumer string
	confirm chan amqp.Confirmation
	signal  chan error
}

// newConn creates new watched AMQP connection
func newConn(url string, tout time.Duration) (*chanPool, error) {
	conn, err := dial(url)
	if err != nil {
		return nil, err
	}

	cp := &chanPool{
		url:       url,
		tout:      tout,
		conn:      conn,
		mu:        &sync.Mutex{},
		channels:  make(map[string]*channel),
		wait:      make(chan interface{}),
		connected: make(chan interface{}),
	}

	close(cp.connected)
	go cp.watch()
	return cp, nil
}

// dial dials to AMQP.
func dial(url string) (*amqp.Connection, error) {
	return amqp.Dial(url)
}

// Close gracefully closes all underlying channels and connection.
func (cp *chanPool) Close() error {
	cp.mu.Lock()

	close(cp.wait)
	if cp.channels == nil {
		return fmt.Errorf("connection is dead")
	}

	// close all channels and consume
	var wg sync.WaitGroup
	for _, ch := range cp.channels {
		wg.Add(1)

		go func(ch *channel) {
			defer wg.Done()
			cp.closeChan(ch, nil)
		}(ch)
	}
	cp.mu.Unlock()

	wg.Wait()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.conn != nil {
		return cp.conn.Close()
	}

	return nil
}

// waitConnected waits till connection is connected again or eventually closed.
// must only be invoked after connection error has been delivered to channel.signal.
func (cp *chanPool) waitConnected() chan interface{} {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return cp.connected
}

// watch manages connection state and reconnects if needed
func (cp *chanPool) watch() {
	for {
		select {
		case <-cp.wait:
			// connection has been closed
			return
			// here we are waiting for the errors from amqp connection
		case err := <-cp.conn.NotifyClose(make(chan *amqp.Error)):
			cp.mu.Lock()
			// clear connected, since connections are dead
			cp.connected = make(chan interface{})

			// broadcast error to all consume to let them for the tryReconnect
			for _, ch := range cp.channels {
				ch.signal <- err
			}

			// disable channel allocation while server is dead
			cp.conn = nil
			cp.channels = nil

			// initialize the backoff
			expb := backoff.NewExponentialBackOff()
			expb.MaxInterval = cp.tout
			cp.mu.Unlock()

			// reconnect function
			reconnect := func() error {
				cp.mu.Lock()
				conn, err := dial(cp.url)
				if err != nil {
					// still failing
					fmt.Println(fmt.Sprintf("error during the amqp dialing, %s", err.Error()))
					cp.mu.Unlock()
					return err
				}

				// TODO ADD LOGGING
				fmt.Println("------amqp successfully redialed------")

				// here we are reconnected
				// replace the connection
				cp.conn = conn
				// re-init the channels
				cp.channels = make(map[string]*channel)
				cp.mu.Unlock()
				return nil
			}

			// start backoff retry
			errb := backoff.Retry(reconnect, expb)
			if errb != nil {
				fmt.Println(fmt.Sprintf("backoff Retry error, %s", errb.Error()))
				// reconnection failed
				close(cp.connected)
				return
			}
			close(cp.connected)
		}
	}
}

// channel allocates new channel on amqp connection
func (cp *chanPool) channel(name string) (*channel, error) {
	cp.mu.Lock()
	dead := cp.conn == nil
	cp.mu.Unlock()

	if dead {
		// wait for connection restoration (doubled the timeout duration)
		select {
		case <-time.NewTimer(cp.tout * 2).C:
			return nil, fmt.Errorf("connection is dead")
		case <-cp.connected:
			// connected
		}
	}

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.conn == nil {
		return nil, fmt.Errorf("connection has been closed")
	}

	if ch, ok := cp.channels[name]; ok {
		return ch, nil
	}

	// we must create new channel
	ch, err := cp.conn.Channel()
	if err != nil {
		return nil, err
	}

	// Enable publish confirmations
	if err = ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("unable to enable confirmation mode on channel: %s", err)
	}

	// we expect that every allocated channel would have listener on signal
	// this is not true only in case of pure producing channels
	cp.channels[name] = &channel{
		ch:      ch,
		confirm: ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
		signal:  make(chan error, 1),
	}

	return cp.channels[name], nil
}

// closeChan gracefully closes and removes channel allocation.
func (cp *chanPool) closeChan(c *channel, err error) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	go func() {
		c.signal <- nil
		c.ch.Close()
	}()

	for name, ch := range cp.channels {
		if ch == c {
			delete(cp.channels, name)
		}
	}

	return err
}

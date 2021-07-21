package beanstalk

import (
	"sync"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/spiral/errors"
)

type ConnPool struct {
	sync.RWMutex
	conn  *beanstalk.Conn
	connT *beanstalk.Conn
	ts    *beanstalk.TubeSet
	t     *beanstalk.Tube

	network string
	address string
	tName   string
	tout    time.Duration
}

func NewConnPool(network, address, tName string, tout time.Duration) (*ConnPool, error) {
	connT, err := beanstalk.DialTimeout(network, address, tout)
	if err != nil {
		return nil, err
	}

	connTS, err := beanstalk.DialTimeout(network, address, tout)
	if err != nil {
		return nil, err
	}

	tube := beanstalk.NewTube(connT, tName)
	ts := beanstalk.NewTubeSet(connTS, tName)

	return &ConnPool{
		network: network,
		address: address,
		tName:   tName,
		tout:    tout,
		conn:    connTS,
		connT:   connT,
		ts:      ts,
		t:       tube,
	}, nil
}

func (cp *ConnPool) Put(body []byte, pri uint32, delay, ttr time.Duration) (uint64, error) {
	cp.RLock()
	defer cp.RUnlock()
	return cp.t.Put(body, pri, delay, ttr)
}

// Reserve reserves and returns a job from one of the tubes in t. If no
// job is available before time timeout has passed, Reserve returns a
// ConnError recording ErrTimeout.
//
// Typically, a client will reserve a job, perform some work, then delete
// the job with Conn.Delete.
func (cp *ConnPool) Reserve(reserveTimeout time.Duration) (id uint64, body []byte, err error) {
	cp.RLock()
	defer cp.RUnlock()
	return cp.ts.Reserve(reserveTimeout)
}

func (cp *ConnPool) Delete(id uint64) error {
	cp.RLock()
	defer cp.RUnlock()
	return cp.conn.Delete(id)
}

func (cp *ConnPool) Redial() error {
	const op = errors.Op("connection_pool_redial")
	connT, err := beanstalk.DialTimeout(cp.network, cp.address, cp.tout)
	if err != nil {
		return err
	}
	if connT == nil {
		return errors.E(op, errors.Str("connectionT is nil"))
	}

	connTS, err := beanstalk.DialTimeout(cp.network, cp.address, cp.tout)
	if err != nil {
		return err
	}

	if connTS == nil {
		return errors.E(op, errors.Str("connectionTS is nil"))
	}

	cp.Lock()
	cp.t = beanstalk.NewTube(connT, cp.tName)
	cp.ts = beanstalk.NewTubeSet(connTS, cp.tName)
	cp.conn = connTS
	cp.connT = connT
	cp.Unlock()
	return nil
}

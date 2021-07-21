package beanstalk

import (
	"strings"
	"sync"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type ConnPool struct {
	sync.RWMutex

	log logger.Logger

	conn  *beanstalk.Conn
	connT *beanstalk.Conn
	ts    *beanstalk.TubeSet
	t     *beanstalk.Tube

	network string
	address string
	tName   string
	tout    time.Duration
}

func NewConnPool(network, address, tName string, tout time.Duration, log logger.Logger) (*ConnPool, error) {
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
		log:     log,
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

	id, err := cp.t.Put(body, pri, delay, ttr)
	if err != nil {
		// errN contains both, err and internal checkAndRedial error
		errN := cp.checkAndRedial(err)
		if errN != nil {
			return 0, errN
		}
	}

	return id, nil
}

// Reserve reserves and returns a job from one of the tubes in t. If no
// job is available before time timeout has passed, Reserve returns a
// ConnError recording ErrTimeout.
//
// Typically, a client will reserve a job, perform some work, then delete
// the job with Conn.Delete.

func (cp *ConnPool) Reserve(reserveTimeout time.Duration) (uint64, []byte, error) {
	cp.RLock()
	defer cp.RUnlock()

	id, body, err := cp.ts.Reserve(reserveTimeout)
	if err != nil {
		errN := cp.checkAndRedial(err)
		if errN != nil {
			return 0, nil, errN
		}

		return 0, nil, err
	}

	return id, body, nil
}

func (cp *ConnPool) Delete(id uint64) error {
	cp.RLock()
	defer cp.RUnlock()

	err := cp.conn.Delete(id)
	if err != nil {
		errN := cp.checkAndRedial(err)
		if errN != nil {
			return errN
		}

		return err
	}
	return nil
}

func (cp *ConnPool) redial() error {
	const op = errors.Op("connection_pool_redial")

	cp.Lock()
	// backoff here
	expb := backoff.NewExponentialBackOff()
	// set the retry timeout (minutes)
	expb.MaxElapsedTime = time.Minute * 5

	operation := func() error {
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

		cp.t = beanstalk.NewTube(connT, cp.tName)
		cp.ts = beanstalk.NewTubeSet(connTS, cp.tName)
		cp.conn = connTS
		cp.connT = connT

		cp.log.Info("beanstalk redial was successful")
		return nil
	}

	retryErr := backoff.Retry(operation, expb)
	if retryErr != nil {
		cp.Unlock()
		return retryErr
	}
	cp.Unlock()

	return nil
}

var connErrors = []string{"pipe", "read tcp", "write tcp", "connection", "EOF"}

func (cp *ConnPool) checkAndRedial(err error) error {
	const op = errors.Op("connection_pool_check_redial")

	for _, errStr := range connErrors {
		if connErr, ok := err.(beanstalk.ConnError); ok {
			// if error is related to the broken connection - redial
			if strings.Contains(errStr, connErr.Err.Error()) {
				cp.RUnlock()
				errR := cp.redial()
				cp.RLock()
				// if redial failed - return
				if errR != nil {
					return errors.E(op, errors.Errorf("%v:%v", err, errR))
				}
				// if redial was successful -> continue listening
				return nil
			}
		}
	}

	return nil
}

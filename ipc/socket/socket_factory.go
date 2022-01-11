package socket

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/pkg/relay"
	"github.com/spiral/goridge/v3/pkg/socket"
	"github.com/spiral/roadrunner/v2/internal"
	"github.com/spiral/roadrunner/v2/worker"
	"go.uber.org/zap"

	"golang.org/x/sync/errgroup"
)

// Factory connects to external stack using socket server.
type Factory struct {
	// listens for incoming connections from underlying processes
	ls net.Listener

	// relay connection timeout
	tout time.Duration

	// sockets which are waiting for process association
	relays sync.Map

	log *zap.Logger
}

// NewSocketServer returns Factory attached to a given socket listener.
// tout specifies for how long factory should serve for incoming relay connection
func NewSocketServer(ls net.Listener, tout time.Duration, log *zap.Logger) *Factory {
	f := &Factory{
		ls:     ls,
		tout:   tout,
		relays: sync.Map{},
		log:    log,
	}

	// Be careful
	// https://github.com/go101/go101/wiki/About-memory-ordering-guarantees-made-by-atomic-operations-in-Go
	// https://github.com/golang/go/issues/5045
	go func() {
		err := f.listen()
		// there is no logger here, use fmt
		if err != nil {
			fmt.Printf("[WARN]: socket server listen, error: %v\n", err)
		}
	}()

	return f
}

// blocking operation, returns an error
func (f *Factory) listen() error {
	errGr := &errgroup.Group{}
	errGr.Go(func() error {
		for {
			conn, err := f.ls.Accept()
			if err != nil {
				return err
			}

			rl := socket.NewSocketRelay(conn)
			pid, err := internal.Pid(rl)
			if err != nil {
				return err
			}

			f.attachRelayToPid(pid, rl)
		}
	})

	return errGr.Wait()
}

type socketSpawn struct {
	w   *worker.Process
	err error
}

// SpawnWorkerWithTimeout creates Process and connects it to appropriate relay or return an error
func (f *Factory) SpawnWorkerWithTimeout(ctx context.Context, cmd *exec.Cmd) (*worker.Process, error) {
	c := make(chan socketSpawn)
	go func() {
		ctxT, cancel := context.WithTimeout(ctx, f.tout)
		defer cancel()
		w, err := worker.InitBaseWorker(cmd, worker.WithLog(f.log))
		if err != nil {
			select {
			case c <- socketSpawn{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		err = w.Start()
		if err != nil {
			select {
			case c <- socketSpawn{
				w:   nil,
				err: err,
			}:
				return
			default:
				return
			}
		}

		rl, err := f.findRelayWithContext(ctxT, w)
		if err != nil {
			_ = w.Kill()
			select {
			// try to write result
			case c <- socketSpawn{
				w:   nil,
				err: err,
			}:
				return
				// if no receivers - return
			default:
				return
			}
		}

		w.AttachRelay(rl)
		w.State().Set(worker.StateReady)

		select {
		case c <- socketSpawn{
			w:   w,
			err: nil,
		}:
			return
		default:
			_ = w.Kill()
			return
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-c:
		if res.err != nil {
			return nil, res.err
		}

		return res.w, nil
	}
}

func (f *Factory) SpawnWorker(cmd *exec.Cmd) (*worker.Process, error) {
	w, err := worker.InitBaseWorker(cmd, worker.WithLog(f.log))
	if err != nil {
		return nil, err
	}

	err = w.Start()
	if err != nil {
		return nil, err
	}

	rl, err := f.findRelay(w)
	if err != nil {
		_ = w.Kill()
		return nil, err
	}

	w.AttachRelay(rl)

	// errors bundle
	_, err = internal.Pid(rl)
	if err != nil {
		_ = w.Kill()
		return nil, err
	}

	w.State().Set(worker.StateReady)

	return w, nil
}

// Close socket factory and underlying socket connection.
func (f *Factory) Close() error {
	return f.ls.Close()
}

// waits for Process to connect over socket and returns associated relay of timeout
func (f *Factory) findRelayWithContext(ctx context.Context, w worker.BaseProcess) (*socket.Relay, error) {
	ticker := time.NewTicker(time.Millisecond * 10)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// check for the process exists
			_, err := process.NewProcess(int32(w.Pid()))
			if err != nil {
				return nil, err
			}
		default:
			// find first pid and attach relay to it
			var r *socket.Relay
			f.relays.Range(func(k, val interface{}) bool {
				r = val.(*socket.Relay)
				f.relays.Delete(k)
				return false
			})

			// no relay exists
			if r == nil {
				continue
			}

			return r, nil
		}
	}
}

func (f *Factory) findRelay(w worker.BaseProcess) (*socket.Relay, error) {
	const op = errors.Op("factory_find_relay")
	// poll every 1ms for the relay
	pollDone := time.NewTimer(f.tout)
	for {
		select {
		case <-pollDone.C:
			return nil, errors.E(op, errors.Str("relay timeout"))
		default:
			tmp, ok := f.relays.Load(w.Pid())
			if !ok {
				continue
			}
			return tmp.(*socket.Relay), nil
		}
	}
}

// chan to store relay associated with specific pid
func (f *Factory) attachRelayToPid(pid int64, relay relay.Relay) {
	f.relays.Store(pid, relay)
}

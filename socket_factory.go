package roadrunner

import (
	"context"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/spiral/errors"

	"github.com/spiral/goridge/v2"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

// SocketFactory connects to external stack using socket server.
type SocketFactory struct {
	// listens for incoming connections from underlying processes
	ls net.Listener

	// relay connection timeout
	tout time.Duration

	// sockets which are waiting for process association
	// relays map[int64]*goridge.SocketRelay
	relays sync.Map

	ErrCh chan error
}

// todo: review

// NewSocketServer returns SocketFactory attached to a given socket listener.
// tout specifies for how long factory should serve for incoming relay connection
func NewSocketServer(ls net.Listener, tout time.Duration) Factory {
	f := &SocketFactory{
		ls:     ls,
		tout:   tout,
		relays: sync.Map{},
		ErrCh:  make(chan error, 10),
	}

	// Be careful
	// https://github.com/go101/go101/wiki/About-memory-ordering-guarantees-made-by-atomic-operations-in-Go
	// https://github.com/golang/go/issues/5045
	go func() {
		f.ErrCh <- f.listen()
	}()

	return f
}

// blocking operation, returns an error
func (f *SocketFactory) listen() error {
	errGr := &errgroup.Group{}
	errGr.Go(func() error {
		for {
			conn, err := f.ls.Accept()
			if err != nil {
				return err
			}

			rl := goridge.NewSocketRelay(conn)
			pid, err := fetchPID(rl)
			if err != nil {
				return err
			}

			f.attachRelayToPid(pid, rl)
		}
	})

	return errGr.Wait()
}

type socketSpawn struct {
	w   WorkerBase
	err error
}

// SpawnWorker creates WorkerProcess and connects it to appropriate relay or returns error
func (f *SocketFactory) SpawnWorkerWithContext(ctx context.Context, cmd *exec.Cmd) (WorkerBase, error) {
	const op = errors.Op("spawn_worker_with_context")
	c := make(chan socketSpawn)
	go func() {
		ctx, cancel := context.WithTimeout(ctx, f.tout)
		defer cancel()
		w, err := InitBaseWorker(cmd)
		if err != nil {
			c <- socketSpawn{
				w:   nil,
				err: err,
			}
			return
		}

		err = w.Start()
		if err != nil {
			c <- socketSpawn{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		rl, err := f.findRelayWithContext(ctx, w)
		if err != nil {
			err = multierr.Combine(
				err,
				w.Kill(),
				w.Wait(context.Background()),
			)

			c <- socketSpawn{
				w:   nil,
				err: errors.E(op, err),
			}
			return
		}

		w.AttachRelay(rl)
		w.State().Set(StateReady)

		c <- socketSpawn{
			w:   w,
			err: nil,
		}
		return
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

func (f *SocketFactory) SpawnWorker(cmd *exec.Cmd) (WorkerBase, error) {
	ctx := context.Background()
	const op = errors.Op("spawn_worker")
	w, err := InitBaseWorker(cmd)
	if err != nil {
		return nil, err
	}

	err = w.Start()
	if err != nil {
		return nil, errors.E(op, err)
	}

	rl, err := f.findRelay(w)
	if err != nil {
		err = multierr.Combine(
			err,
			w.Kill(),
			w.Wait(ctx),
		)
		return nil, err
	}

	w.AttachRelay(rl)
	w.State().Set(StateReady)

	return w, nil
}

// Close socket factory and underlying socket connection.
func (f *SocketFactory) Close(ctx context.Context) error {
	return f.ls.Close()
}

// waits for WorkerProcess to connect over socket and returns associated relay of timeout
func (f *SocketFactory) findRelayWithContext(ctx context.Context, w WorkerBase) (*goridge.SocketRelay, error) {
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			_, err := process.NewProcess(int32(w.Pid()))
			if err != nil {
				return nil, err
			}
		default:
			tmp, ok := f.relays.Load(w.Pid())
			if !ok {
				continue
			}
			return tmp.(*goridge.SocketRelay), nil
		}
	}
}

func (f *SocketFactory) findRelay(w WorkerBase) (*goridge.SocketRelay, error) {
	const op = errors.Op("find_relay")
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
			return tmp.(*goridge.SocketRelay), nil
		}
	}
}

// chan to store relay associated with specific pid
func (f *SocketFactory) attachRelayToPid(pid int64, relay *goridge.SocketRelay) {
	f.relays.Store(pid, relay)
}

// deletes relay chan associated with specific pid
func (f *SocketFactory) removeRelayFromPid(pid int64) {
	f.relays.Delete(pid)
}

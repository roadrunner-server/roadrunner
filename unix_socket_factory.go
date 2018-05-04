package roadrunner

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spiral/goridge"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

// UnixSocketFactory connects to external workers using socket server.
type UnixSocketFactory struct {
	// relay connection timeout
	tout time.Duration

	// protects socket mapping
	mu sync.Mutex

	// unix socket file
	file string

	// sockets which are waiting for process association
	relays map[int]chan *goridge.SocketRelay
}

// NewUnixSocketFactory returns UnixSocketFactory attached to a given socket listener.
// tout specifies for how long factory should serve for incoming relay connection
func NewUnixSocketFactory(file string, tout time.Duration) *UnixSocketFactory {
	f := &UnixSocketFactory{
		file:   file,
		tout:   tout,
		relays: make(map[int]chan *goridge.SocketRelay),
	}

	return f
}

// SpawnWorker creates worker and connects it to appropriate relay or returns error
func (f *UnixSocketFactory) SpawnWorker(cmd *exec.Cmd) (w *Worker, err error) {
	if w, err = newWorker(cmd); err != nil {
		return nil, err
	}

	if err := w.start(); err != nil {
		return nil, errors.Wrap(err, "process error")
	}

	socketFile := fmt.Sprintf("%s.%d", f.file, *w.Pid)

	os.Remove(socketFile)
	ls, err := net.Listen("unix", socketFile)
	if err != nil {
		return nil, err
	}

	go f.listen(ls)

	rl, err := f.findRelay(w, f.tout)
	if err != nil {
		go func(w *Worker) { w.Kill() }(w)

		if wErr := w.Wait(); wErr != nil {
			err = errors.Wrap(wErr, err.Error())
		}

		return nil, errors.Wrap(err, "unable to connect to worker")
	}

	w.rl = rl
	w.state.set(StateReady)

	return w, nil
}

// listens for incoming socket connection only a time
func (f *UnixSocketFactory) listen(ls net.Listener) {
	conn, err := ls.Accept()
	if err != nil {
		return
	}

	defer ls.Close()

	rl := goridge.NewSocketRelay(conn)
	if pid, err := fetchPID(rl); err == nil {
		f.relayChan(pid) <- rl
	}
}

// waits for worker to connect over socket and returns associated relay of timeout
func (f *UnixSocketFactory) findRelay(w *Worker, tout time.Duration) (*goridge.SocketRelay, error) {
	timer := time.NewTimer(tout)
	for {
		select {
		case rl := <-f.relayChan(*w.Pid):
			timer.Stop()
			f.cleanChan(*w.Pid)
			return rl, nil

		case <-timer.C:
			return nil, fmt.Errorf("relay timeout")

		case <-w.waitDone:
			timer.Stop()
			f.cleanChan(*w.Pid)
			return nil, fmt.Errorf("worker is gone")
		}
	}
}

// chan to store relay associated with specific Pid
func (f *UnixSocketFactory) relayChan(pid int) chan *goridge.SocketRelay {
	f.mu.Lock()
	defer f.mu.Unlock()

	rl, ok := f.relays[pid]
	if !ok {
		f.relays[pid] = make(chan *goridge.SocketRelay)
		return f.relays[pid]
	}

	return rl
}

// deletes relay chan associated with specific Pid
func (f *UnixSocketFactory) cleanChan(pid int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.relays, pid)
}

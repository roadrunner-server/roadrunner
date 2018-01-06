package roadrunner

import (
	"fmt"
	"github.com/spiral/goridge"
	"net"
	"os/exec"
	"sync"
	"time"
)

// SocketFactory connects to external workers using socket server.
type SocketFactory struct {
	ls   net.Listener                      // listens for incoming connections from underlying processes
	tout time.Duration                     // connection timeout
	mu   sync.Mutex                        // protects socket mapping
	wait map[int]chan *goridge.SocketRelay // sockets which are waiting for process association
}

// NewSocketFactory returns SocketFactory attached to a given socket listener. tout specifies for how long factory
// should wait for incoming relay connection
func NewSocketFactory(ls net.Listener, tout time.Duration) *SocketFactory {
	f := &SocketFactory{
		ls:   ls,
		tout: tout,
		wait: make(map[int]chan *goridge.SocketRelay),
	}

	go f.listen()
	return f
}

// NewWorker creates worker and connects it to appropriate relay or returns error
func (f *SocketFactory) NewWorker(cmd *exec.Cmd) (w *Worker, err error) {
	w, err = newWorker(cmd)
	if err != nil {
		return nil, err
	}

	if err := w.Start(); err != nil {
		return nil, err
	}

	rl, err := f.waitRelay(*w.Pid, f.tout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to worker: %s", err)
	}

	w.attach(rl)
	w.st = newState(StateReady)

	return w, nil
}

// Close closes all open factory descriptors.
func (f *SocketFactory) Close() error {
	return f.ls.Close()
}

// listen for incoming wait and associate sockets with active workers
func (f *SocketFactory) listen() {
	for {
		conn, err := f.ls.Accept()
		if err != nil {
			return
		}

		rl := goridge.NewSocketRelay(conn)
		if pid, err := fetchPid(rl); err == nil {
			f.relayChan(pid) <- rl
		}
	}
}

// waits for worker to connect over socket and returns associated relay of timeout
func (f *SocketFactory) waitRelay(pid int, tout time.Duration) (*goridge.SocketRelay, error) {
	timer := time.NewTimer(tout)
	select {
	case rl := <-f.relayChan(pid):
		timer.Stop()
		f.cleanChan(pid)

		return rl, nil
	case <-timer.C:
		return nil, fmt.Errorf("relay timeout")
	}
}

// chan to store relay associated with specific Pid
func (f *SocketFactory) relayChan(pid int) chan *goridge.SocketRelay {
	f.mu.Lock()
	defer f.mu.Unlock()

	rl, ok := f.wait[pid]
	if !ok {
		f.wait[pid] = make(chan *goridge.SocketRelay)
		return f.wait[pid]
	}

	return rl
}

// deletes relay chan associated with specific Pid
func (f *SocketFactory) cleanChan(pid int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.wait, pid)
}

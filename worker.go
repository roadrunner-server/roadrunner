package roadrunner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/util"

	"github.com/spiral/goridge/v3"
	"go.uber.org/multierr"
)

const (
	// WaitDuration - for how long error buffer should attempt to aggregate error messages
	// before merging output together since lastError update (required to keep error update together).
	WaitDuration = 25 * time.Millisecond

	// ReadBufSize used to make a slice with specified length to read from stderr
	ReadBufSize = 10240 // Kb
)

// EventWorkerKill thrown after WorkerProcess is being forcefully killed.
const (
	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError Event = iota + 200

	// EventWorkerLog triggered on every write to WorkerProcess StdErr pipe (batched). Except payload to be []byte string.
	EventWorkerLog
)

type Event int64

func (ev Event) String() string {
	switch ev {
	case EventWorkerError:
		return "EventWorkerError"
	case EventWorkerLog:
		return "EventWorkerLog"
	}
	return "Unknown event type"
}

// WorkerEvent wraps worker events.
type WorkerEvent struct {
	// Event id, see below.
	Event Event

	// Worker triggered the event.
	Worker WorkerBase

	// Event specific payload.
	Payload interface{}
}

var pool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, ReadBufSize)
		return &buf
	},
}

type WorkerBase interface {
	fmt.Stringer

	// Pid returns worker pid.
	Pid() int64

	// Created returns time worker was created at.
	Created() time.Time

	// AddListener attaches listener to consume worker events.
	AddListener(listener util.EventListener)

	// State return receive-only WorkerProcess state object, state can be used to safely access
	// WorkerProcess status, time when status changed and number of WorkerProcess executions.
	State() State

	// Start used to run Cmd and immediately return
	Start() error

	// Wait must be called once for each WorkerProcess, call will be released once WorkerProcess is
	// complete and will return process error (if any), if stderr is presented it's value
	// will be wrapped as WorkerError. Method will return error code if php process fails
	// to find or Start the script.
	Wait() error

	// Stop sends soft termination command to the WorkerProcess and waits for process completion.
	Stop(ctx context.Context) error

	// Kill kills underlying process, make sure to call Wait() func to gather
	// error log from the stderr. Does not waits for process completion!
	Kill() error

	// Relay returns attached to worker goridge relay
	Relay() goridge.Relay

	// AttachRelay used to attach goridge relay to the worker process
	AttachRelay(rl goridge.Relay)
}

// WorkerProcess - supervised process with api over goridge.Relay.
type WorkerProcess struct {
	// created indicates at what time WorkerProcess has been created.
	created time.Time

	// updates parent supervisor or pool about WorkerProcess events
	events util.EventsHandler

	// state holds information about current WorkerProcess state,
	// number of WorkerProcess executions, buf status change time.
	// publicly this object is receive-only and protected using Mutex
	// and atomic counter.
	state *state

	// underlying command with associated process, command must be
	// provided to WorkerProcess from outside in non-started form. CmdSource
	// stdErr direction will be handled by WorkerProcess to aggregate error message.
	cmd *exec.Cmd

	// pid of the process, points to pid of underlying process and
	// can be nil while process is not started.
	pid int

	// stderr aggregates stderr output from underlying process. Value can be
	// receive only once command is completed and all pipes are closed.
	stderr *bytes.Buffer

	// channel is being closed once command is complete.
	// waitDone chan interface{}

	// contains information about resulted process state.
	endState *os.ProcessState

	// ensures than only one execution can be run at once.
	mu sync.RWMutex

	// communication bus with underlying process.
	relay goridge.Relay
	rd    io.Reader
	stop  chan struct{}
}

// InitBaseWorker creates new WorkerProcess over given exec.cmd.
func InitBaseWorker(cmd *exec.Cmd) (WorkerBase, error) {
	if cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}
	w := &WorkerProcess{
		created: time.Now(),
		events:  &util.EventHandler{},
		cmd:     cmd,
		state:   newState(StateInactive),
		stderr:  new(bytes.Buffer),
		stop:    make(chan struct{}, 1),
	}

	w.rd, w.cmd.Stderr = io.Pipe()

	// small buffer optimization
	// at this point we know, that stderr will contain huge messages
	w.stderr.Grow(ReadBufSize)

	go func() {
		w.watch()
	}()

	return w, nil
}

// Pid returns worker pid.
func (w *WorkerProcess) Pid() int64 {
	return int64(w.pid)
}

// Created returns time worker was created at.
func (w *WorkerProcess) Created() time.Time {
	return w.created
}

// AddListener registers new worker event listener.
func (w *WorkerProcess) AddListener(listener util.EventListener) {
	w.events.AddListener(listener)
}

// State return receive-only WorkerProcess state object, state can be used to safely access
// WorkerProcess status, time when status changed and number of WorkerProcess executions.
func (w *WorkerProcess) State() State {
	return w.state
}

// State return receive-only WorkerProcess state object, state can be used to safely access
// WorkerProcess status, time when status changed and number of WorkerProcess executions.
func (w *WorkerProcess) AttachRelay(rl goridge.Relay) {
	w.relay = rl
}

// State return receive-only WorkerProcess state object, state can be used to safely access
// WorkerProcess status, time when status changed and number of WorkerProcess executions.
func (w *WorkerProcess) Relay() goridge.Relay {
	return w.relay
}

// String returns WorkerProcess description. fmt.Stringer interface
func (w *WorkerProcess) String() string {
	st := w.state.String()
	// we can safely compare pid to 0
	if w.pid != 0 {
		st = st + ", pid:" + strconv.Itoa(w.pid)
	}

	return fmt.Sprintf(
		"(`%s` [%s], numExecs: %v)",
		strings.Join(w.cmd.Args, " "),
		st,
		w.state.NumExecs(),
	)
}

func (w *WorkerProcess) Start() error {
	err := w.cmd.Start()
	if err != nil {
		return err
	}
	w.pid = w.cmd.Process.Pid
	return nil
}

// Wait must be called once for each WorkerProcess, call will be released once WorkerProcess is
// complete and will return process error (if any), if stderr is presented it's value
// will be wrapped as WorkerError. Method will return error code if php process fails
// to find or Start the script.
func (w *WorkerProcess) Wait() error {
	const op = errors.Op("worker process wait")
	err := multierr.Combine(w.cmd.Wait())

	// at this point according to the documentation (see cmd.Wait comment)
	// if worker finishes with an error, message will be written to the stderr first
	// and then w.cmd.Wait return an error
	w.endState = w.cmd.ProcessState
	if err != nil {
		w.state.Set(StateErrored)

		w.mu.RLock()
		// if process return code > 0, here will be an error from stderr (if presents)
		if w.stderr.Len() > 0 {
			err = multierr.Append(err, errors.E(op, errors.Str(w.stderr.String())))
			// stop the stderr buffer
			w.stop <- struct{}{}
		}
		w.mu.RUnlock()

		return multierr.Append(err, w.closeRelay())
	}

	err = multierr.Append(err, w.closeRelay())
	if err != nil {
		w.state.Set(StateErrored)
		return err
	}

	if w.endState.Success() {
		w.state.Set(StateStopped)
	}

	return nil
}

func (w *WorkerProcess) closeRelay() error {
	if w.relay != nil {
		err := w.relay.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop sends soft termination command to the WorkerProcess and waits for process completion.
func (w *WorkerProcess) Stop(ctx context.Context) error {
	c := make(chan error)

	go func() {
		var err error
		w.state.Set(StateStopping)
		err = multierr.Append(err, sendControl(w.relay, &stopCommand{Stop: true}))
		if err != nil {
			w.state.Set(StateKilling)
			c <- multierr.Append(err, w.cmd.Process.Kill())
		}
		w.state.Set(StateStopped)
		c <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		if err != nil {
			return err
		}
		return nil
	}
}

// Kill kills underlying process, make sure to call Wait() func to gather
// error log from the stderr. Does not waits for process completion!
func (w *WorkerProcess) Kill() error {
	w.state.Set(StateKilling)
	err := w.cmd.Process.Signal(os.Kill)
	if err != nil {
		return err
	}
	w.state.Set(StateStopped)
	return nil
}

// put the pointer, to not allocate new slice
// but erase it len and then return back
func (w *WorkerProcess) put(data *[]byte) {
	*data = (*data)[:0]
	*data = (*data)[:cap(*data)]

	pool.Put(data)
}

// get pointer to the byte slice
func (w *WorkerProcess) get() *[]byte {
	return pool.Get().(*[]byte)
}

// Write appends the contents of pool to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of pool; errBuffer is always nil.
func (w *WorkerProcess) watch() {
	go func() {
		for {
			select {
			case <-w.stop:
				buf := w.get()
				// read the last data
				n, _ := w.rd.Read(*buf)
				w.events.Push(WorkerEvent{Event: EventWorkerLog, Worker: w, Payload: (*buf)[:n]})
				w.mu.Lock()
				// write new message
				w.stderr.Write((*buf)[:n])
				w.mu.Unlock()
				w.put(buf)
				return
			default:
				// read the max 10kb of stderr per one read
				buf := w.get()
				n, _ := w.rd.Read(*buf)
				w.events.Push(WorkerEvent{Event: EventWorkerLog, Worker: w, Payload: (*buf)[:n]})
				w.mu.Lock()
				// write new message
				w.stderr.Write((*buf)[:n])
				w.mu.Unlock()
				w.put(buf)
			}
		}
	}()
}

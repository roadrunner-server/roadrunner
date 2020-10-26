package roadrunner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spiral/roadrunner/v2/util"

	"github.com/spiral/goridge/v2"
	"go.uber.org/multierr"
)

const (
	// WaitDuration - for how long error buffer should attempt to aggregate error messages
	// before merging output together since lastError update (required to keep error update together).
	WaitDuration = 25 * time.Millisecond
)

// EventWorkerKill thrown after WorkerProcess is being forcefully killed.
const (
	// EventWorkerError triggered after WorkerProcess. Except payload to be error.
	EventWorkerError int64 = iota + 200

	// EventWorkerLog triggered on every write to WorkerProcess StdErr pipe (batched). Except payload to be []byte string.
	EventWorkerLog
)

// WorkerEvent wraps worker events.
type WorkerEvent struct {
	// Event id, see below.
	Event int64

	// Worker triggered the event.
	Worker WorkerBase

	// Event specific payload.
	Payload interface{}
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
	Wait(ctx context.Context) error

	// Stop sends soft termination command to the WorkerProcess and waits for process completion.
	Stop(ctx context.Context) error

	// Kill kills underlying process, make sure to call Wait() func to gather
	// error log from the stderr. Does not waits for process completion!
	Kill(ctx context.Context) error

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
	events *util.EventHandler

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

	// errBuffer aggregates stderr output from underlying process. Value can be
	// receive only once command is completed and all pipes are closed.
	errBuffer *errBuffer

	// channel is being closed once command is complete.
	// waitDone chan interface{}

	// contains information about resulted process state.
	endState *os.ProcessState

	// ensures than only one execution can be run at once.
	mu sync.Mutex

	// communication bus with underlying process.
	relay goridge.Relay
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
	}

	w.errBuffer = newErrBuffer(w.logCallback)

	// piping all stderr to command errBuffer
	w.cmd.Stderr = w.errBuffer

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

	w.errBuffer.mu.Lock()
	w.errBuffer.enable = true
	w.errBuffer.mu.Unlock()
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
func (w *WorkerProcess) Wait(ctx context.Context) error {
	err := multierr.Combine(w.cmd.Wait())
	w.endState = w.cmd.ProcessState
	if err != nil {
		w.state.Set(StateErrored)

		// if no errors in the events, error might be in the errBuffer
		if w.errBuffer.Len() > 0 {
			err = multierr.Append(err, errors.New(w.errBuffer.String()))
		}

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
		w.errBuffer.Close()
		w.state.Set(StateStopping)
		w.mu.Lock()
		defer w.mu.Unlock()
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
func (w *WorkerProcess) Kill(ctx context.Context) error {
	w.state.Set(StateKilling)
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.cmd.Process.Signal(os.Kill)
	if err != nil {
		return err
	}
	w.state.Set(StateStopped)
	return nil
}

func (w *WorkerProcess) logCallback(log []byte) {
	w.events.Push(WorkerEvent{Event: EventWorkerLog, Worker: w, Payload: log})
}

// thread safe errBuffer
type errBuffer struct {
	enable bool
	mu     sync.RWMutex
	buf    []byte
	last   int
	wait   *time.Timer
	// todo: remove update
	update      chan interface{}
	stop        chan interface{}
	logCallback func(log []byte)
}

func newErrBuffer(logCallback func(log []byte)) *errBuffer {
	eb := &errBuffer{
		buf:         make([]byte, 0),
		update:      make(chan interface{}),
		wait:        time.NewTimer(WaitDuration),
		stop:        make(chan interface{}),
		logCallback: logCallback,
	}

	go func(eb *errBuffer) {
		for {
			select {
			case <-eb.update:
				eb.wait.Reset(WaitDuration)
			case <-eb.wait.C:
				eb.mu.Lock()
				if eb.enable && len(eb.buf) > eb.last {
					eb.logCallback(eb.buf[eb.last:])
					eb.buf = eb.buf[0:0]
					eb.last = len(eb.buf)
				}
				eb.mu.Unlock()
			case <-eb.stop:
				eb.wait.Stop()

				eb.mu.Lock()
				if eb.enable && len(eb.buf) > eb.last {
					eb.logCallback(eb.buf[eb.last:])
					eb.last = len(eb.buf)
				}
				eb.mu.Unlock()
				return
			}
		}
	}(eb)

	return eb
}

// Len returns the number of buf of the unread portion of the errBuffer;
// buf.Len() == len(buf.Bytes()).
func (eb *errBuffer) Len() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// currently active message
	return len(eb.buf)
}

// Write appends the contents of pool to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of pool; errBuffer is always nil.
func (eb *errBuffer) Write(p []byte) (int, error) {
	eb.mu.Lock()
	eb.buf = append(eb.buf, p...)
	eb.mu.Unlock()
	eb.update <- nil

	return len(p), nil
}

// Strings fetches all errBuffer data into string.
func (eb *errBuffer) String() string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// TODO unsafe operation, use runes
	return string(eb.buf)
}

// Close aggregation timer.
func (eb *errBuffer) Close() {
	close(eb.stop)
}

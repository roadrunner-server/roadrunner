package worker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/interfaces/relay"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/interfaces/worker"
	"github.com/spiral/roadrunner/v2/internal"
	eventsPkg "github.com/spiral/roadrunner/v2/pkg/events"
	"go.uber.org/multierr"
)

const (
	// WaitDuration - for how long error buffer should attempt to aggregate error messages
	// before merging output together since lastError update (required to keep error update together).
	WaitDuration = 25 * time.Millisecond

	// ReadBufSize used to make a slice with specified length to read from stderr
	ReadBufSize = 10240 // Kb
)

type Options func(p *Process)

// Process - supervised process with api over goridge.Relay.
type Process struct {
	// created indicates at what time Process has been created.
	created time.Time

	// updates parent supervisor or pool about Process events
	events events.Handler

	// state holds information about current Process state,
	// number of Process executions, buf status change time.
	// publicly this object is receive-only and protected using Mutex
	// and atomic counter.
	state *internal.WorkerState

	// underlying command with associated process, command must be
	// provided to Process from outside in non-started form. CmdSource
	// stdErr direction will be handled by Process to aggregate error message.
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
	relay relay.Relay
	// rd in a second part of pipe to read from stderr
	rd io.Reader
	// stop signal terminates io.Pipe from reading from stderr
	stop chan struct{}

	syncPool sync.Pool
}

// InitBaseWorker creates new Process over given exec.cmd.
func InitBaseWorker(cmd *exec.Cmd, options ...Options) (worker.BaseProcess, error) {
	if cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}
	w := &Process{
		created: time.Now(),
		events:  eventsPkg.NewEventsHandler(),
		cmd:     cmd,
		state:   internal.NewWorkerState(internal.StateInactive),
		stderr:  new(bytes.Buffer),
		stop:    make(chan struct{}, 1),
		// sync pool for STDERR
		// All receivers are pointers
		syncPool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, ReadBufSize)
				return &buf
			},
		},
	}

	w.rd, w.cmd.Stderr = io.Pipe()

	// small buffer optimization
	// at this point we know, that stderr will contain huge messages
	w.stderr.Grow(ReadBufSize)

	// add options
	for i := 0; i < len(options); i++ {
		options[i](w)
	}

	go func() {
		w.watch()
	}()

	return w, nil
}

func AddListeners(listeners ...events.EventListener) Options {
	return func(p *Process) {
		for i := 0; i < len(listeners); i++ {
			p.addListener(listeners[i])
		}
	}
}

// Pid returns worker pid.
func (w *Process) Pid() int64 {
	return int64(w.pid)
}

// Created returns time worker was created at.
func (w *Process) Created() time.Time {
	return w.created
}

// AddListener registers new worker event listener.
func (w *Process) addListener(listener events.EventListener) {
	w.events.AddListener(listener)
}

// State return receive-only Process state object, state can be used to safely access
// Process status, time when status changed and number of Process executions.
func (w *Process) State() internal.State {
	return w.state
}

// State return receive-only Process state object, state can be used to safely access
// Process status, time when status changed and number of Process executions.
func (w *Process) AttachRelay(rl relay.Relay) {
	w.relay = rl
}

// State return receive-only Process state object, state can be used to safely access
// Process status, time when status changed and number of Process executions.
func (w *Process) Relay() relay.Relay {
	return w.relay
}

// String returns Process description. fmt.Stringer interface
func (w *Process) String() string {
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

func (w *Process) Start() error {
	err := w.cmd.Start()
	if err != nil {
		return err
	}
	w.pid = w.cmd.Process.Pid
	return nil
}

// Wait must be called once for each Process, call will be released once Process is
// complete and will return process error (if any), if stderr is presented it's value
// will be wrapped as WorkerError. Method will return error code if php process fails
// to find or Start the script.
func (w *Process) Wait() error {
	const op = errors.Op("worker process wait")
	err := multierr.Combine(w.cmd.Wait())

	// at this point according to the documentation (see cmd.Wait comment)
	// if worker finishes with an error, message will be written to the stderr first
	// and then w.cmd.Wait return an error
	w.endState = w.cmd.ProcessState
	if err != nil {
		w.state.Set(internal.StateErrored)

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
		w.state.Set(internal.StateErrored)
		return err
	}

	if w.endState.Success() {
		w.state.Set(internal.StateStopped)
	}

	return nil
}

func (w *Process) closeRelay() error {
	if w.relay != nil {
		err := w.relay.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop sends soft termination command to the Process and waits for process completion.
func (w *Process) Stop() error {
	var err error
	w.state.Set(internal.StateStopping)
	err = multierr.Append(err, internal.SendControl(w.relay, &internal.StopCommand{Stop: true}))
	if err != nil {
		w.state.Set(internal.StateKilling)
		return multierr.Append(err, w.cmd.Process.Kill())
	}
	w.state.Set(internal.StateStopped)
	return nil
}

// Kill kills underlying process, make sure to call Wait() func to gather
// error log from the stderr. Does not waits for process completion!
func (w *Process) Kill() error {
	w.state.Set(internal.StateKilling)
	err := w.cmd.Process.Signal(os.Kill)
	if err != nil {
		return err
	}
	w.state.Set(internal.StateStopped)
	return nil
}

// put the pointer, to not allocate new slice
// but erase it len and then return back
func (w *Process) put(data *[]byte) {
	w.syncPool.Put(data)
}

// get pointer to the byte slice
func (w *Process) get() *[]byte {
	return w.syncPool.Get().(*[]byte)
}

// Write appends the contents of pool to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of pool; errBuffer is always nil.
func (w *Process) watch() {
	go func() {
		for {
			select {
			case <-w.stop:
				buf := w.get()
				// read the last data
				n, _ := w.rd.Read(*buf)
				w.events.Push(events.WorkerEvent{Event: events.EventWorkerLog, Worker: w, Payload: (*buf)[:n]})
				w.mu.Lock()
				// write new message
				// we are sending only n read bytes, without sending previously written message as bytes slice from syncPool
				w.stderr.Write((*buf)[:n])
				w.mu.Unlock()
				w.put(buf)
				return
			default:
				// read the max 10kb of stderr per one read
				buf := w.get()
				n, _ := w.rd.Read(*buf)
				w.events.Push(events.WorkerEvent{Event: events.EventWorkerLog, Worker: w, Payload: (*buf)[:n]})
				w.mu.Lock()
				// write new message
				w.stderr.Write((*buf)[:n])
				w.mu.Unlock()
				w.put(buf)
			}
		}
	}()
}

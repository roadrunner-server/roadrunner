package worker

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/goridge/v3/interfaces/relay"
	"github.com/spiral/roadrunner/v2/internal"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/states"
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

	// contains information about resulted process state.
	endState *os.ProcessState

	// communication bus with underlying process.
	relay relay.Relay
}

// InitBaseWorker creates new Process over given exec.cmd.
func InitBaseWorker(cmd *exec.Cmd, options ...Options) (*Process, error) {
	const op = errors.Op("init_base_worker")
	if cmd.Process != nil {
		return nil, fmt.Errorf("can't attach to running process")
	}
	w := &Process{
		created: time.Now(),
		events:  events.NewEventsHandler(),
		cmd:     cmd,
		state:   internal.NewWorkerState(states.StateInactive),
	}

	// set self as stderr implementation (Writer interface)
	w.cmd.Stderr = w

	// add options
	for i := 0; i < len(options); i++ {
		options[i](w)
	}

	return w, nil
}

func AddListeners(listeners ...events.Listener) Options {
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
func (w *Process) addListener(listener events.Listener) {
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
	const op = errors.Op("process_wait")
	var err error
	err = w.cmd.Wait()

	// If worker was destroyed, just exit
	if w.State().Value() == states.StateDestroyed {
		return nil
	}

	// If state is different, and err is not nil, append it to the errors
	if err != nil {
		w.State().Set(states.StateErrored)
		err = multierr.Combine(err, errors.E(op, err))
	}

	// closeRelay
	// at this point according to the documentation (see cmd.Wait comment)
	// if worker finishes with an error, message will be written to the stderr first
	// and then process.cmd.Wait return an error
	err2 := w.closeRelay()
	if err2 != nil {
		w.State().Set(states.StateErrored)
		return multierr.Append(err, errors.E(op, err2))
	}

	if w.cmd.ProcessState.Success() {
		w.State().Set(states.StateStopped)
		return nil
	}

	return err
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
	w.state.Set(states.StateStopping)
	err = multierr.Append(err, internal.SendControl(w.relay, &internal.StopCommand{Stop: true}))
	if err != nil {
		w.state.Set(states.StateKilling)
		return multierr.Append(err, w.cmd.Process.Kill())
	}
	w.state.Set(states.StateStopped)
	return nil
}

// Kill kills underlying process, make sure to call Wait() func to gather
// error log from the stderr. Does not waits for process completion!
func (w *Process) Kill() error {
	if w.State().Value() == states.StateDestroyed {
		err := w.cmd.Process.Signal(os.Kill)
		if err != nil {
			return err
		}
		return nil
	}

	w.state.Set(states.StateKilling)
	err := w.cmd.Process.Signal(os.Kill)
	if err != nil {
		return err
	}
	w.state.Set(states.StateStopped)
	return nil
}

// Worker stderr
func (w *Process) Write(p []byte) (n int, err error) {
	w.events.Push(events.WorkerEvent{Event: events.EventWorkerStderr, Worker: w, Payload: p})
	return len(p), nil
}

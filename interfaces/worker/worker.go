package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/spiral/goridge/v3"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/internal"
)

// Allocator is responsible for worker allocation in the pool
type Allocator func() (BaseProcess, error)

type BaseProcess interface {
	fmt.Stringer

	// Pid returns worker pid.
	Pid() int64

	// Created returns time worker was created at.
	Created() time.Time

	// AddListener attaches listener to consume worker events.
	AddListener(listener events.EventListener)

	// State return receive-only WorkerProcess state object, state can be used to safely access
	// WorkerProcess status, time when status changed and number of WorkerProcess executions.
	State() internal.State

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

type SyncWorker interface {
	// BaseProcess provides basic functionality for the SyncWorker
	BaseProcess
	// Exec used to execute payload on the SyncWorker, there is no TIMEOUTS
	Exec(rqs internal.Payload) (internal.Payload, error)
	// ExecWithContext used to handle Exec with TTL
	ExecWithContext(ctx context.Context, p internal.Payload) (internal.Payload, error)
}

package service

import (
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// Process structure contains an information about process, restart information, log, errors, etc
type Process struct {
	sync.Mutex
	// command to execute
	command *exec.Cmd
	rawCmd  string

	errCh chan error
	log   logger.Logger

	ExecTimeout      time.Duration
	RestartAfterExit bool
	RestartDelay     time.Duration

	//
	startTime time.Time
	stopCh    chan struct{}
}

// NewServiceProcess constructs service process structure
func NewServiceProcess(restartAfterExit bool, execTimeout, restartDelay time.Duration, command string, l logger.Logger, errCh chan error) *Process {
	return &Process{
		rawCmd:           command,
		RestartDelay:     restartDelay,
		ExecTimeout:      execTimeout,
		RestartAfterExit: restartAfterExit,
		errCh:            errCh,
		stopCh:           make(chan struct{}),
		log:              l,
	}
}

// write message to the log (stderr)
func (p *Process) Write(b []byte) (int, error) {
	p.log.Info(toString(b))
	return len(b), nil
}

func (p *Process) start() {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("processor_start")

	// crate fat-process here
	p.createProcess()

	err := p.command.Start()
	if err != nil {
		p.errCh <- errors.E(op, err)
		return
	}

	go p.wait()
	go p.execHandler()
	// save start time
	p.startTime = time.Now()
}

func (p *Process) createProcess() {
	// cmdArgs contain command arguments if the command in form of: php <command> or ls <command> -i -b
	var cmdArgs []string
	cmdArgs = append(cmdArgs, strings.Split(p.rawCmd, " ")...)
	if len(cmdArgs) < 2 {
		p.command = exec.Command(p.rawCmd) //nolint:gosec
	} else {
		p.command = exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
	}
	p.command.Stderr = p
}

func (p *Process) wait() {
	// Wait error doesn't matter here
	_ = p.command.Wait()

	// wait for restart delay
	if p.RestartAfterExit {
		// wait for the delay
		time.Sleep(p.RestartDelay)
		// and start command again
		p.start()
	}
}

// stop can be only sent by the Endure when plugin stopped
func (p *Process) stop() {
	p.stopCh <- struct{}{}
}

func (p *Process) execHandler() {
	tt := time.NewTicker(time.Second)
	for {
		select {
		case <-tt.C:
			// lock here, because p.startTime could be changed during the check
			p.Lock()
			// if the exec timeout is set
			if p.ExecTimeout != 0 {
				// check the running time for the script
				if time.Now().After(p.startTime.Add(p.ExecTimeout)) {
					err := p.command.Process.Signal(syscall.SIGINT)
					if err != nil {
						_ = p.command.Process.Signal(syscall.SIGKILL)
					}
				}
			}
			p.Unlock()
		case <-p.stopCh:
			err := p.command.Process.Signal(syscall.SIGINT)
			if err != nil {
				_ = p.command.Process.Signal(syscall.SIGKILL)
			}
			tt.Stop()
			return
		}
	}
}

// unsafe and fast []byte to string convert
//go:inline
func toString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

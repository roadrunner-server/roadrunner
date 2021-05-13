package service

import (
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/utils"
)

// Process structure contains an information about process, restart information, log, errors, etc
type Process struct {
	sync.Mutex
	// command to execute
	command *exec.Cmd
	// rawCmd from the plugin
	rawCmd string
	Pid    int

	// root plugin error chan
	errCh chan error
	// logger
	log logger.Logger

	ExecTimeout     time.Duration
	RemainAfterExit bool
	RestartSec      uint64

	// process start time
	startTime time.Time
	stopped   uint64
}

// NewServiceProcess constructs service process structure
func NewServiceProcess(restartAfterExit bool, execTimeout time.Duration, restartDelay uint64, command string, l logger.Logger, errCh chan error) *Process {
	return &Process{
		rawCmd:          command,
		RestartSec:      restartDelay,
		ExecTimeout:     execTimeout,
		RemainAfterExit: restartAfterExit,
		errCh:           errCh,
		log:             l,
	}
}

// write message to the log (stderr)
func (p *Process) Write(b []byte) (int, error) {
	p.log.Info(utils.AsString(b))
	return len(b), nil
}

func (p *Process) start() {
	p.Lock()
	defer p.Unlock()
	const op = errors.Op("processor_start")

	// crate fat-process here
	p.createProcess()

	// non blocking process start
	err := p.command.Start()
	if err != nil {
		p.errCh <- errors.E(op, err)
		return
	}

	// start process waiting routine
	go p.wait()
	// execHandler checks for the execTimeout
	go p.execHandler()
	// save start time
	p.startTime = time.Now()
	p.Pid = p.command.Process.Pid
}

// create command for the process
func (p *Process) createProcess() {
	// cmdArgs contain command arguments if the command in form of: php <command> or ls <command> -i -b
	var cmdArgs []string
	cmdArgs = append(cmdArgs, strings.Split(p.rawCmd, " ")...)
	if len(cmdArgs) < 2 {
		p.command = exec.Command(p.rawCmd) //nolint:gosec
	} else {
		p.command = exec.Command(cmdArgs[0], cmdArgs[1:]...) //nolint:gosec
	}
	// redirect stderr into the Write function of the process.go
	p.command.Stderr = p
}

// wait process for exit
func (p *Process) wait() {
	// Wait error doesn't matter here
	err := p.command.Wait()
	if err != nil {
		p.log.Error("process wait error", "error", err)
	}
	// wait for restart delay
	if p.RemainAfterExit {
		// wait for the delay
		time.Sleep(time.Second * time.Duration(p.RestartSec))
		// and start command again
		p.start()
	}
}

// stop can be only sent by the Endure when plugin stopped
func (p *Process) stop() {
	atomic.StoreUint64(&p.stopped, 1)
}

func (p *Process) execHandler() {
	tt := time.NewTicker(time.Second)
	for range tt.C {
		// lock here, because p.startTime could be changed during the check
		p.Lock()
		// if the exec timeout is set
		if p.ExecTimeout != 0 {
			// if stopped -> kill the process (SIGINT-> SIGKILL) and exit
			if atomic.CompareAndSwapUint64(&p.stopped, 1, 1) {
				err := p.command.Process.Signal(syscall.SIGINT)
				if err != nil {
					_ = p.command.Process.Signal(syscall.SIGKILL)
				}
				tt.Stop()
				p.Unlock()
				return
			}

			// check the running time for the script
			if time.Now().After(p.startTime.Add(p.ExecTimeout)) {
				err := p.command.Process.Signal(syscall.SIGINT)
				if err != nil {
					_ = p.command.Process.Signal(syscall.SIGKILL)
				}
				p.Unlock()
				tt.Stop()
				return
			}
		}
		p.Unlock()
	}
}

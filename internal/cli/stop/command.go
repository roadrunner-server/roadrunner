package stop

import (
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const (
	// sync with root.go
	pidFileName string = ".pid"
)

// NewCommand creates `serve` command.
func NewCommand(silent *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop RoadRunner server",
		RunE: func(*cobra.Command, []string) error {
			const op = errors.Op("rr_stop")

			data, err := os.ReadFile(pidFileName)
			if err != nil {
				return errors.Errorf("%v, to create a .pid file, you must run RR with the following options: './rr serve -p'", err)
			}

			pid, err := strconv.Atoi(string(data))
			if err != nil {
				return errors.E(op, err)
			}

			process, err := os.FindProcess(pid)
			if err != nil {
				return errors.E(op, err)
			}

			if !*silent {
				log.Printf("stopping process with PID: %d", pid)
			}

			err = process.Signal(syscall.SIGTERM)
			if err != nil {
				return errors.E(op, err)
			}

			return nil
		},
	}
}

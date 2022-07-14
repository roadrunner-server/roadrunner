package serve

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/roadrunner-server/roadrunner/v2/roadrunner"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const (
	rrPrefix string = "rr"
)

// NewCommand creates `serve` command.
func NewCommand(override *[]string, cfgFile *string, silent *bool) *cobra.Command { //nolint:funlen
	return &cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner server",
		RunE: func(*cobra.Command, []string) error {
			const op = errors.Op("handle_serve_command")
			rr, err := roadrunner.NewRR(*cfgFile, override, roadrunner.DefaultPluginsList())

			if err != nil {
				return errors.E(op, err)
			}

			errCh := make(chan error, 1)
			go func() {
				err = rr.Serve()
				if err != nil {
					errCh <- errors.E(op, err)
				}
			}()

			oss, stop := make(chan os.Signal, 5), make(chan struct{}, 1) //nolint:gomnd
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

			go func() {
				// first catch - stop the container
				<-oss
				// send signal to stop execution
				stop <- struct{}{}

				// after first hit we are waiting for the second
				// second catch - exit from the process
				<-oss
				fmt.Println("exit forced")
				os.Exit(1)
			}()

			for {
				select {
				case e := <-errCh:
					return errors.E(op, e)
				case <-stop: // stop the container after first signal
					fmt.Printf("stop signal received\n")

					if err = rr.Stop(); err != nil {
						return fmt.Errorf("error: %w", err)
					}

					return nil
				}
			}
		},
	}
}

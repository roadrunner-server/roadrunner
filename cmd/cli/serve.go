package cli

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	"go.uber.org/multierr"
)

func init() {
	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner server",
		RunE:  handler,
	})
}

func handler(_ *cobra.Command, _ []string) error {
	const op = errors.Op("handle_serve_command")
	/*
		We need to have path to the config at the RegisterTarget stage
		But after cobra.Execute, because cobra fills up cli variables on this stage
	*/

	err := Container.Init()
	if err != nil {
		return errors.E(op, err)
	}

	errCh, err := Container.Serve()
	if err != nil {
		return errors.E(op, err)
	}

	// https://golang.org/pkg/os/signal/#Notify
	// should be of buffer size at least 1
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case e := <-errCh:
			err = multierr.Append(err, e.Error)
			log.Printf("error occurred: %v, service: %s", e.Error.Error(), e.VertexID)
			er := Container.Stop()
			if er != nil {
				err = multierr.Append(err, er)
				return errors.E(op, err)
			}
			return errors.E(op, err)
		case <-c:
			err = Container.Stop()
			if err != nil {
				return errors.E(op, err)
			}
			return nil
		}
	}
}

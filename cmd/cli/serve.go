package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spiral/errors"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
)

func init() {
	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner Temporal service(s)",
		RunE:  handler,
	})
}

func handler(cmd *cobra.Command, args []string) error {
	const op = errors.Op("handle serve command")
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
			Logger.Error(e.Error.Error(), zap.String("service", e.VertexID))
			er := Container.Stop()
			if er != nil {
				Logger.Error(e.Error.Error(), zap.String("service", e.VertexID))
				if er != nil {
					return errors.E(op, er)
				}
			}
		case <-c:
			err = Container.Stop()
			if err != nil {
				return errors.E(op, err)
			}
			return nil
		}
	}
}

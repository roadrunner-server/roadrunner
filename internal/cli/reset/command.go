package reset

import (
	"log"
	"sync"

	internalRpc "github.com/roadrunner-server/roadrunner/v2/internal/rpc"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const (
	op            = errors.Op("reset_handler")
	resetterList  = "resetter.List"
	resetterReset = "resetter.Reset"
)

// NewCommand creates `reset` command.
func NewCommand(cfgFile *string, override *[]string, silent *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset workers of all or specific RoadRunner service",
		RunE: func(_ *cobra.Command, args []string) error {
			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			client, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			defer func() { _ = client.Close() }()

			plugins := args        // by default we expect services list from user
			if len(plugins) == 0 { // but if nothing was passed - request all services list
				if err = client.Call(resetterList, true, &plugins); err != nil {
					return err
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(plugins))

			for _, plugin := range plugins {
				// simulating some work
				go func(p string) {
					if !*silent {
						log.Printf("resetting plugin: [%s] ", p)
					}
					defer wg.Done()

					var done bool
					<-client.Go(resetterReset, p, &done, nil).Done

					if err != nil {
						log.Println(err)

						return
					}

					if !*silent {
						log.Printf("plugin reset: [%s]", p)
					}
				}(plugin)
			}

			wg.Wait()

			return nil
		},
	}
}

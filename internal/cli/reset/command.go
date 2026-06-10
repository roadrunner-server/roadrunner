package reset

import (
	"log"
	"sync"

	internalRpc "github.com/roadrunner-server/roadrunner/v2025/internal/rpc"
	"github.com/roadrunner-server/roadrunner/v2025/internal/sdnotify"

	"connectrpc.com/connect"
	resetterV1 "github.com/roadrunner-server/api-go/v6/resetter/v1"
	"github.com/roadrunner-server/api-go/v6/resetter/v1/resetterV1connect"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const op = errors.Op("reset_handler")

// NewCommand creates `reset` command.
func NewCommand(cfgFile *string, override *[]string, silent *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset workers of all or specific RoadRunner service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			baseURL, httpClient, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			client := resetterV1connect.NewResetterServiceClient(httpClient, baseURL)
			ctx := cmd.Context()

			plugins := args        // by default, we expect services list from user
			if len(plugins) == 0 { // but if nothing was passed - request all services list
				resp, errL := client.ListPlugins(ctx, connect.NewRequest(&resetterV1.ListPluginsRequest{}))
				if errL != nil {
					return errL
				}

				plugins = resp.Msg.GetPlugins()
			}

			_, _ = sdnotify.SdNotify(sdnotify.Reloading)

			var wg sync.WaitGroup
			for _, plugin := range plugins {
				wg.Go(func() {
					if !*silent {
						log.Printf("resetting plugin: [%s] ", plugin)
					}

					_, errR := client.Reset(ctx, connect.NewRequest(&resetterV1.ResetRequest{Plugin: plugin}))
					if errR != nil {
						log.Println(errR)

						return
					}

					if !*silent {
						log.Printf("plugin reset: [%s]", plugin)
					}
				})
			}

			wg.Wait()

			_, _ = sdnotify.SdNotify(sdnotify.Ready)

			return nil
		},
	}
}

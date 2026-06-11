package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalRpc "github.com/roadrunner-server/roadrunner/v3/internal/rpc"

	"connectrpc.com/connect"
	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	informerV1 "github.com/roadrunner-server/api-go/v6/informer/v1"
	"github.com/roadrunner-server/api-go/v6/informer/v1/informerV1connect"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

// NewCommand creates `workers` command.
func NewCommand(cfgFile *string, override *[]string) *cobra.Command { //nolint:funlen
	// interactive workers updates
	var interactive bool

	cmd := &cobra.Command{
		Use:   "workers",
		Short: "Show information about active RoadRunner workers",
		RunE: func(cmd *cobra.Command, args []string) error {
			const op = errors.Op("handle_workers_command")

			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			baseURL, httpClient, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			client := informerV1connect.NewInformerServiceClient(httpClient, baseURL)
			ctx := cmd.Context()

			plugins := args        // by default, we expect a plugin list from user
			if len(plugins) == 0 { // but if nothing was passed - request all informers list
				resp, errL := client.ListPlugins(ctx, connect.NewRequest(&informerV1.ListPluginsRequest{}))
				if errL != nil {
					return fmt.Errorf("failed to get list of plugins: %w", errL)
				}

				plugins = resp.Msg.GetPlugins()
			}

			if !interactive {
				showWorkers(ctx, plugins, client)
				return nil
			}

			oss := make(chan os.Signal, 1)
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

			tm.Clear()

			tt := time.NewTicker(time.Second)
			defer tt.Stop()

			for {
				select {
				case <-oss:
					return nil

				case <-tt.C:
					tm.MoveCursor(1, 1)
					tm.Flush()

					showWorkers(ctx, plugins, client)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(
		&interactive,
		"interactive",
		"i",
		false,
		"render interactive workers table",
	)

	return cmd
}

func showWorkers(ctx context.Context, plugins []string, client informerV1connect.InformerServiceClient) {
	// this is only one exception to Render the workers, service plugin has the same workers as other plugins,
	// but they are RAW processes and needs to be handled in a different way. We don't need a special RPC call, but
	// need a special render method.
	const servicePluginName = "service"

	for _, plugin := range plugins {
		resp, err := client.GetWorkers(ctx, connect.NewRequest(&informerV1.GetWorkersRequest{Plugin: plugin}))
		if err != nil {
			// this is a special case, when we can't get workers list, we need to render an error message
			_ = WorkerTable(os.Stdout, nil, fmt.Errorf("failed to receive information about %s plugin: %w", plugin, err)).Render()
			continue
		}

		workers := resp.Msg.GetWorkers()
		if len(workers) == 0 {
			continue
		}

		fmt.Printf("Workers of [%s]:\n", color.HiYellowString(plugin))

		if plugin == servicePluginName {
			_ = ServiceWorkerTable(os.Stdout, workers).Render()
			continue
		}

		_ = WorkerTable(os.Stdout, workers, nil).Render()
	}

	for _, plugin := range plugins {
		resp, err := client.GetJobs(ctx, connect.NewRequest(&informerV1.GetJobsRequest{Plugin: plugin}))
		if err != nil {
			_ = JobsTable(os.Stdout, nil, fmt.Errorf("failed to receive information about %s plugin: %w", plugin, err)).Render()
			continue
		}

		jst := resp.Msg.GetStates()
		if len(jst) == 0 {
			continue
		}

		fmt.Printf("Jobs of [%s]:\n", color.HiYellowString(plugin))
		_ = JobsTable(os.Stdout, jst, nil).Render()
	}
}

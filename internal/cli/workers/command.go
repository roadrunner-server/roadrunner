package workers

import (
	"fmt"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roadrunner-server/api/v4/plugins/v4/jobs"
	internalRpc "github.com/roadrunner-server/roadrunner/v2024/internal/rpc"

	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/informer/v5"
	"github.com/spf13/cobra"
)

// NewCommand creates `workers` command.
func NewCommand(cfgFile *string, override *[]string) *cobra.Command { //nolint:funlen
	// interactive workers updates
	var interactive bool

	cmd := &cobra.Command{
		Use:   "workers",
		Short: "Show information about active RoadRunner workers",
		RunE: func(_ *cobra.Command, args []string) error {
			const (
				op           = errors.Op("handle_workers_command")
				informerList = "informer.List"
			)

			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			client, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			defer func() { _ = client.Close() }()

			plugins := args        // by default, we expect a plugin list from user
			if len(plugins) == 0 { // but if nothing was passed - request all informers list
				if err = client.Call(informerList, true, &plugins); err != nil {
					return fmt.Errorf("failed to get list of plugins: %w", err)
				}
			}

			if !interactive {
				showWorkers(plugins, client)
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

					showWorkers(plugins, client)
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

func showWorkers(plugins []string, client *rpc.Client) {
	const (
		informerWorkers = "informer.Workers"
		informerJobs    = "informer.Jobs"
		// this is only one exception to Render the workers, service plugin has the same workers as other plugins,
		// but they are RAW processes and needs to be handled in a different way. We don't need a special RPC call, but
		// need a special render method.
		servicePluginName = "service"
	)

	for _, plugin := range plugins {
		list := &informer.WorkerList{}

		if err := client.Call(informerWorkers, plugin, &list); err != nil {
			// this is a special case, when we can't get workers list, we need to render an error message
			WorkerTable(os.Stdout, list.Workers, fmt.Errorf("failed to receive information about %s plugin: %w", plugin, err)).Render()
			continue
		}

		if len(list.Workers) == 0 {
			continue
		}

		if plugin == servicePluginName {
			fmt.Printf("Workers of [%s]:\n", color.HiYellowString(plugin))
			ServiceWorkerTable(os.Stdout, list.Workers).Render()

			continue
		}

		fmt.Printf("Workers of [%s]:\n", color.HiYellowString(plugin))

		WorkerTable(os.Stdout, list.Workers, nil).Render()
	}

	for _, plugin := range plugins {
		var jst []*jobs.State

		if err := client.Call(informerJobs, plugin, &jst); err != nil {
			JobsTable(os.Stdout, jst, fmt.Errorf("failed to receive information about %s plugin: %w", plugin, err)).Render()
			continue
		}

		// eq to nil
		if len(jst) == 0 {
			continue
		}

		fmt.Printf("Jobs of [%s]:\n", color.HiYellowString(plugin))
		JobsTable(os.Stdout, jst, nil).Render()
	}
}

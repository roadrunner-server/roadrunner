package workers

import (
	"fmt"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roadrunner-server/api/v2/plugins/jobs"
	internalRpc "github.com/spiral/roadrunner-binary/v2/internal/rpc"

	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/informer"
)

// NewCommand creates `workers` command.
func NewCommand(cfgPlugin *config.Plugin) *cobra.Command { //nolint:funlen
	var (
		// interactive workers updates
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "workers",
		Short: "Show information about active RoadRunner workers",
		RunE: func(_ *cobra.Command, args []string) error {
			const (
				op           = errors.Op("handle_workers_command")
				informerList = "informer.List"
			)

			client, err := internalRpc.NewClient(cfgPlugin)
			if err != nil {
				return err
			}

			defer func() { _ = client.Close() }()

			plugins := args        // by default we expect plugins list from user
			if len(plugins) == 0 { // but if nothing was passed - request all informers list
				if err = client.Call(informerList, true, &plugins); err != nil {
					return err
				}
			}

			if !interactive {
				return showWorkers(plugins, client)
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

					if err = showWorkers(plugins, client); err != nil {
						return errors.E(op, err)
					}
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

func showWorkers(plugins []string, client *rpc.Client) error {
	const (
		op              = errors.Op("show_workers")
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
			return errors.E(op, err)
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

		WorkerTable(os.Stdout, list.Workers).Render()
	}

	for _, plugin := range plugins {
		var jst []*jobs.State

		if err := client.Call(informerJobs, plugin, &jst); err != nil {
			return errors.E(op, err)
		}

		// eq to nil
		if len(jst) == 0 {
			continue
		}

		fmt.Printf("Jobs of [%s]:\n", color.HiYellowString(plugin))
		JobsTable(os.Stdout, jst).Render()
	}

	return nil
}

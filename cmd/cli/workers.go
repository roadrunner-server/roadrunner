package cli

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/tools"
)

// use interactive mode
var interactive bool

const InformerList string = "informer.List"
const InformerWorkers string = "informer.Workers"

func init() {
	workersCommand := &cobra.Command{
		Use:   "workers",
		Short: "Show information about active roadrunner workers",
		RunE:  workersHandler,
	}

	workersCommand.Flags().BoolVarP(
		&interactive,
		"interactive",
		"i",
		false,
		"render interactive workers table",
	)

	root.AddCommand(workersCommand)
}

func workersHandler(_ *cobra.Command, args []string) error {
	const op = errors.Op("handle_workers_command")
	// get RPC client
	client, err := RPCClient()
	if err != nil {
		return err
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Printf("error when closing RPCClient: error %v", err)
		}
	}()

	var plugins []string
	// assume user wants to show workers from particular plugin
	if len(args) != 0 {
		plugins = args
	} else {
		err = client.Call(InformerList, true, &plugins)
		if err != nil {
			return errors.E(op, err)
		}
	}

	if !interactive {
		return showWorkers(plugins, client)
	}

	// https://golang.org/pkg/os/signal/#Notify
	// should be of buffer size at least 1
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	tm.Clear()
	tt := time.NewTicker(time.Second)
	defer tt.Stop()
	for {
		select {
		case <-c:
			return nil
		case <-tt.C:
			tm.MoveCursor(1, 1)
			tm.Flush()
			err := showWorkers(plugins, client)
			if err != nil {
				return errors.E(op, err)
			}
		}
	}
}

func showWorkers(plugins []string, client *rpc.Client) error {
	const op = errors.Op("show_workers")
	for _, plugin := range plugins {
		list := &informer.WorkerList{}
		err := client.Call(InformerWorkers, plugin, &list)
		if err != nil {
			return errors.E(op, err)
		}

		fmt.Printf("Workers of [%s]:\n", color.HiYellowString(plugin))
		tools.WorkerTable(os.Stdout, list.Workers).Render()
	}
	return nil
}

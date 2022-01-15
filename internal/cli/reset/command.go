package reset

import (
	"fmt"
	"sync"

	internalRpc "github.com/spiral/roadrunner-binary/v2/internal/rpc"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

var spinnerStyle = []string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"} //nolint:gochecknoglobals

// NewCommand creates `reset` command.
func NewCommand(cfgPlugin *config.Plugin) *cobra.Command { //nolint:funlen
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset workers of all or specific RoadRunner service",
		RunE: func(_ *cobra.Command, args []string) error {
			const (
				op            = errors.Op("reset_handler")
				resetterList  = "resetter.List"
				resetterReset = "resetter.Reset"
			)

			client, err := internalRpc.NewClient(cfgPlugin)
			if err != nil {
				return err
			}

			defer func() { _ = client.Close() }()

			services := args        // by default we expect services list from user
			if len(services) == 0 { // but if nothing was passed - request all services list
				if err = client.Call(resetterList, true, &services); err != nil {
					return err
				}
			}

			var wg sync.WaitGroup
			wg.Add(len(services))

			pr := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(6)) //nolint:gomnd

			for _, service := range services {
				var (
					bar    *mpb.Bar
					name   = runewidth.FillRight(fmt.Sprintf("Resetting plugin: [%s]", color.HiYellowString(service)), 27)
					result = make(chan interface{})
				)

				bar = pr.AddSpinner(
					1,
					mpb.SpinnerOnMiddle,
					mpb.SpinnerStyle(spinnerStyle),
					mpb.PrependDecorators(decor.Name(name)),
					mpb.AppendDecorators(onComplete(result)),
				)

				// simulating some work
				go func(service string, result chan interface{}) {
					defer wg.Done()
					defer bar.Increment()

					var done bool
					<-client.Go(resetterReset, service, &done, nil).Done

					if err != nil {
						result <- errors.E(op, err)

						return
					}

					result <- nil
				}(service, result)
			}

			pr.Wait()

			return nil
		},
	}
}

func onComplete(result chan interface{}) decor.Decorator {
	return decor.Any(func(s decor.Statistics) string {
		select {
		case r := <-result:
			if err, ok := r.(error); ok {
				return color.HiRedString(err.Error())
			}

			return color.HiGreenString("done")
		default:
			return ""
		}
	})
}

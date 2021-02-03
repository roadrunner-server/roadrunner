package cli

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
)

// List is the resetter.List RPC method
const List string = "resetter.List"

// Reset is the resetter.Reset RPC method
const Reset string = "resetter.Reset"

func init() {
	root.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset workers of all or specific RoadRunner service",
		RunE:  resetHandler,
	})
}

func resetHandler(_ *cobra.Command, args []string) error {
	const op = errors.Op("reset_handler")
	client, err := RPCClient()
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	var services []string
	if len(args) != 0 {
		services = args
	} else {
		err = client.Call(List, true, &services)
		if err != nil {
			return errors.E(op, err)
		}
	}

	var wg sync.WaitGroup
	pr := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithWidth(6))
	wg.Add(len(services))

	for _, service := range services {
		var (
			bar    *mpb.Bar
			name   = runewidth.FillRight(fmt.Sprintf("Resetting plugin: [%s]", color.HiYellowString(service)), 27)
			result = make(chan interface{})
		)

		bar = pr.AddSpinner(
			1,
			mpb.SpinnerOnMiddle,
			mpb.SpinnerStyle([]string{"∙∙∙", "●∙∙", "∙●∙", "∙∙●", "∙∙∙"}),
			mpb.PrependDecorators(decor.Name(name)),
			mpb.AppendDecorators(onComplete(result)),
		)

		// simulating some work
		go func(service string, result chan interface{}) {
			defer wg.Done()
			defer bar.Increment()

			var done bool
			cl := client.Go(Reset, service, &done, nil)
			<-cl.Done
			if err != nil {
				result <- errors.E(op, err)
				return
			}
			result <- nil
		}(service, result)
	}

	pr.Wait()
	return nil
}

func onComplete(result chan interface{}) decor.Decorator {
	var (
		msg = ""
		fn  = func(s decor.Statistics) string {
			select {
			case r := <-result:
				if err, ok := r.(error); ok {
					msg = color.HiRedString(err.Error())
					return msg
				}

				msg = color.HiGreenString("done")
				return msg
			default:
				return msg
			}
		}
	)

	return decor.Any(fn)
}

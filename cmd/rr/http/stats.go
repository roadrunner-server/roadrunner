// Copyright (c) 2018 SpiralScout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package http

import (
	tm "github.com/buger/goterm"
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/cmd/util"
	"github.com/spiral/roadrunner/service/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	STATS_interactive bool
	STATS_stopSignal  = make(chan os.Signal, 1)
)

func init() {
	command := &cobra.Command{
		Use:   "http:stats",
		Short: "Stats for RoadRunner HTTP service",
		RunE:  statsHandler,
	}

	command.Flags().BoolVarP(
		&STATS_interactive,
		"interactive",
		"i",
		false,
		"render interactive stats table",
	)

	rr.CLI.AddCommand(command)

	signal.Notify(STATS_stopSignal, syscall.SIGTERM)
	signal.Notify(STATS_stopSignal, syscall.SIGINT)
}

func statsHandler(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if r, ok := recover().(error); ok {
			err = r
		}
	}()

	client, err := util.RPCClient(rr.Container)
	if err != nil {
		return err
	}
	defer client.Close()

	if !STATS_interactive {
		showStats(client)
		return nil
	}

	tm.Clear()
	for {
		select {
		case <-STATS_stopSignal:
			return nil
		case <-time.NewTicker(time.Millisecond * 500).C:
			tm.MoveCursor(1, 1)
			showStats(client)
			tm.Flush()
		}
	}
}

func showStats(client *rpc.Client) {
	var r http.StatsResponse

	if err := client.Call("http.Stats", true, &r); err != nil {
		panic(err)
	}

	util.StatsTable(r.Stats).Render()
}

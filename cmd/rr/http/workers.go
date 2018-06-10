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
	"errors"
	"github.com/spf13/cobra"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/rpc"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/http"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"time"
	"github.com/dustin/go-humanize"
	"github.com/spiral/roadrunner/cmd/rr/utils"
	"github.com/shirou/gopsutil/process"
)

func init() {
	rr.CLI.AddCommand(&cobra.Command{
		Use:   "http:workers",
		Short: "List workers associated with RoadRunner HTTP service",
		RunE:  workersHandler,
	})
}

func workersHandler(cmd *cobra.Command, args []string) error {
	svc, st := rr.Container.Get(rpc.Name)
	if st < service.StatusConfigured {
		return errors.New("RPC service is not configured")
	}

	client, err := svc.(*rpc.Service).Client()
	if err != nil {
		return err
	}
	defer client.Close()

	var r http.WorkerList
	if err := client.Call("http.Workers", true, &r); err != nil {
		panic(err)
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"PID", "Status", "Execs", "Memory", "Created"})

	for _, w := range r.Workers {
		tw.Append([]string{
			strconv.Itoa(w.Pid),
			renderStatus(w.Status),
			renderJobs(w.NumJobs),
			renderMemory(w.Pid),
			renderAlive(time.Unix(0, w.Created)),
		})
	}

	tw.Render()

	return nil
}

func renderStatus(status string) string {
	switch status {
	case "inactive":
		return utils.Sprintf("<yellow>inactive</reset>")
	case "ready":
		return utils.Sprintf("<cyan>ready</reset>")
	case "working":
		return utils.Sprintf("<green>working</reset>")
	case "stopped":
		return utils.Sprintf("<red>stopped</reset>")
	case "errored":
		return utils.Sprintf("<red>errored</reset>")
	}

	return status
}

func renderJobs(number uint64) string {
	return humanize.Comma(int64(number))
}

func renderAlive(t time.Time) string {
	return humanize.RelTime(t, time.Now(), "ago", "")
}

func renderMemory(pid int) string {
	p, _ := process.NewProcess(int32(pid))
	i, err := p.MemoryInfo()
	if err != nil {
		return err.Error()
	}

	return humanize.Bytes(i.RSS)
}

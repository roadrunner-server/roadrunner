package util

import (
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	rrutil "github.com/spiral/roadrunner/util"
	"os"
	"strconv"
	"time"
)

// WorkerTable renders table with information about rr server workers.
func WorkerTable(workers []*rrutil.State) *tablewriter.Table {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader([]string{"PID", "Status", "Execs", "Memory", "Created"})
	tw.SetColMinWidth(0, 7)
	tw.SetColMinWidth(1, 9)
	tw.SetColMinWidth(2, 7)
	tw.SetColMinWidth(3, 7)
	tw.SetColMinWidth(4, 18)

	for _, w := range workers {
		tw.Append([]string{
			strconv.Itoa(w.Pid),
			renderStatus(w.Status),
			renderJobs(w.NumJobs),
			humanize.Bytes(w.MemoryUsage),
			renderAlive(time.Unix(0, w.Created)),
		})
	}

	return tw
}

func renderStatus(status string) string {
	switch status {
	case "inactive":
		return Sprintf("<yellow>inactive</reset>")
	case "ready":
		return Sprintf("<cyan>ready</reset>")
	case "working":
		return Sprintf("<green>working</reset>")
	case "invalid":
		return Sprintf("<yellow>invalid</reset>")
	case "stopped":
		return Sprintf("<red>stopped</reset>")
	case "errored":
		return Sprintf("<red>errored</reset>")
	}

	return status
}

func renderJobs(number int64) string {
	return humanize.Comma(int64(number))
}

func renderAlive(t time.Time) string {
	return humanize.RelTime(t, time.Now(), "ago", "")
}

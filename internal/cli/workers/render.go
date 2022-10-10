package workers

import (
	"io"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/roadrunner-server/sdk/v3/plugins/jobs"
	"github.com/roadrunner-server/sdk/v3/state/process"
)

const (
	Ready  string = "READY"
	Paused string = "PAUSED/STOPPED"
)

// WorkerTable renders table with information about rr server workers.
func WorkerTable(writer io.Writer, workers []*process.State) *tablewriter.Table {
	tw := tablewriter.NewWriter(writer)
	tw.SetHeader([]string{"PID", "Status", "Execs", "Memory", "CPU%", "Created"})
	tw.SetColMinWidth(0, 7)
	tw.SetColMinWidth(1, 9)
	tw.SetColMinWidth(2, 7)
	tw.SetColMinWidth(3, 7)
	tw.SetColMinWidth(4, 7)
	tw.SetColMinWidth(5, 18)

	for i := 0; i < len(workers); i++ {
		tw.Append([]string{
			strconv.Itoa(int(workers[i].Pid)),
			renderStatus(workers[i].StatusStr),
			renderJobs(workers[i].NumExecs),
			humanize.Bytes(workers[i].MemoryUsage),
			renderCPU(workers[i].CPUPercent),
			renderAlive(time.Unix(0, workers[i].Created)),
		})
	}

	return tw
}

// ServiceWorkerTable renders table with information about rr server workers.
func ServiceWorkerTable(writer io.Writer, workers []*process.State) *tablewriter.Table {
	tw := tablewriter.NewWriter(writer)
	tw.SetAutoWrapText(false)
	tw.SetHeader([]string{"PID", "Memory", "CPU%", "Command"})
	tw.SetColMinWidth(0, 7)
	tw.SetColMinWidth(1, 7)
	tw.SetColMinWidth(2, 7)
	tw.SetColMinWidth(3, 18)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)

	for i := 0; i < len(workers); i++ {
		tw.Append([]string{
			strconv.Itoa(int(workers[i].Pid)),
			humanize.Bytes(workers[i].MemoryUsage),
			renderCPU(workers[i].CPUPercent),
			workers[i].Command,
		})
	}

	return tw
}

// JobsTable renders table with information about rr server jobs.
func JobsTable(writer io.Writer, jobs []*jobs.State) *tablewriter.Table {
	tw := tablewriter.NewWriter(writer)
	tw.SetAutoWrapText(false)
	tw.SetHeader([]string{"Status", "Pipeline", "Driver", "Queue", "Active", "Delayed", "Reserved"})
	tw.SetColWidth(10)
	tw.SetColWidth(10)
	tw.SetColWidth(7)
	tw.SetColWidth(15)
	tw.SetColWidth(10)
	tw.SetColWidth(10)
	tw.SetColWidth(10)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)

	for i := 0; i < len(jobs); i++ {
		tw.Append([]string{
			renderReady(jobs[i].Ready),
			jobs[i].Pipeline,
			jobs[i].Driver,
			jobs[i].Queue,
			strconv.Itoa(int(jobs[i].Active)),
			strconv.Itoa(int(jobs[i].Delayed)),
			strconv.Itoa(int(jobs[i].Reserved)),
		})
	}

	return tw
}

func renderReady(ready bool) string {
	if ready {
		return Ready
	}

	return Paused
}

//go:inline
func renderCPU(cpu float64) string {
	return strconv.FormatFloat(cpu, 'f', 2, 64)
}

func renderStatus(status string) string {
	switch status {
	case "inactive":
		return color.YellowString("inactive")
	case "ready":
		return color.CyanString("ready")
	case "working":
		return color.GreenString("working")
	case "invalid":
		return color.YellowString("invalid")
	case "stopped":
		return color.RedString("stopped")
	case "errored":
		return color.RedString("errored")
	default:
		return status
	}
}

func renderJobs(number uint64) string {
	return humanize.Comma(int64(number))
}

func renderAlive(t time.Time) string {
	return humanize.RelTime(t, time.Now(), "ago", "")
}

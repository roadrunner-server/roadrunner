package workers

import (
	"io"
	"sort"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	informerV1 "github.com/roadrunner-server/api-go/v6/informer/v1"
)

const (
	Ready  string = "READY"
	Paused string = "PAUSED/STOPPED"
)

// WorkerTable renders table with information about rr server workers.
func WorkerTable(writer io.Writer, workers []*informerV1.ProcessState, err error) *tablewriter.Table {
	cfg := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoFormat: tw.On,
			},
		},
		MaxWidth: 150,
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{
				Global: tw.AlignLeft,
			},
		},
	}
	tw := tablewriter.NewTable(writer, tablewriter.WithConfig(cfg))
	tw.Header([]string{"PID", "Status", "Execs", "Memory", "CPU%", "Created"})

	if err != nil {
		_ = tw.Append([]string{
			"0",
			err.Error(),
			"ERROR",
			"ERROR",
			"ERROR",
			"ERROR",
		})

		return tw
	}

	sort.Slice(workers, func(i, j int) bool {
		return workers[i].GetPid() < workers[j].GetPid()
	})

	for i := range workers {
		_ = tw.Append([]string{
			strconv.Itoa(int(workers[i].GetPid())),
			renderStatus(workers[i].GetStatusStr()),
			renderJobs(workers[i].GetNumExecs()),
			humanize.Bytes(workers[i].GetMemoryUsage()),
			renderCPU(float64(workers[i].GetCpuPercent())),
			renderAlive(time.Unix(0, workers[i].GetCreated())),
		})
	}

	return tw
}

// ServiceWorkerTable renders table with information about rr server workers.
func ServiceWorkerTable(writer io.Writer, workers []*informerV1.ProcessState) *tablewriter.Table {
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].GetPid() < workers[j].GetPid()
	})

	cfg := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoFormat: tw.On,
			},
		},
		MaxWidth: 150,
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{
				Global: tw.AlignLeft,
			},
		},
	}
	tw := tablewriter.NewTable(writer, tablewriter.WithConfig(cfg))
	tw.Header([]string{"PID", "Memory", "CPU%", "Command"})

	for i := range workers {
		_ = tw.Append([]string{
			strconv.Itoa(int(workers[i].GetPid())),
			humanize.Bytes(workers[i].GetMemoryUsage()),
			renderCPU(float64(workers[i].GetCpuPercent())),
			workers[i].GetCommand(),
		})
	}

	return tw
}

// JobsTable renders table with information about rr server jobs.
func JobsTable(writer io.Writer, jobs []*informerV1.JobState, err error) *tablewriter.Table {
	cfg := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoFormat: tw.On,
				AutoWrap:   int(tw.Off),
			},
		},
		MaxWidth: 150,
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{
				Global: tw.AlignLeft,
			},
		},
	}
	tw := tablewriter.NewTable(writer, tablewriter.WithConfig(cfg))
	tw.Header([]string{"Status", "Pipeline", "Driver", "Queue", "Active", "Delayed", "Reserved"})

	if err != nil {
		_ = tw.Append([]string{
			err.Error(),
			"ERROR",
			"ERROR",
			"ERROR",
			"ERROR",
			"ERROR",
			"ERROR",
		})

		return tw
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].GetPipeline() < jobs[j].GetPipeline()
	})

	for i := range jobs {
		_ = tw.Append([]string{
			renderReady(jobs[i].GetReady()),
			jobs[i].GetPipeline(),
			jobs[i].GetDriver(),
			jobs[i].GetQueue(),
			strconv.Itoa(int(jobs[i].GetActive())),
			strconv.Itoa(int(jobs[i].GetDelayed())),
			strconv.Itoa(int(jobs[i].GetReserved())),
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
	return humanize.Comma(int64(number)) //nolint:gosec
}

func renderAlive(t time.Time) string {
	return humanize.RelTime(t, time.Now(), "ago", "")
}

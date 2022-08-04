package jobs

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// JobsCommandsRender uses console renderer to show jobs
func renderPipelines(writer io.Writer, pipelines []string) *tablewriter.Table {
	tw := tablewriter.NewWriter(writer)
	tw.SetAutoWrapText(false)
	tw.SetHeader([]string{"Pipeline(s)"})
	tw.SetColWidth(50)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)

	for i := 0; i < len(pipelines); i++ {
		tw.Append([]string{pipelines[i]})
	}

	return tw
}

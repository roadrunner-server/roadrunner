package jobs

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// JobsCommandsRender uses console renderer to show jobs
func renderPipelines(writer io.Writer, pipelines []string) *tablewriter.Table {
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
	tw.Header([]string{"Pipeline(s)"})

	for i := range pipelines {
		_ = tw.Append([]string{pipelines[i]})
	}

	return tw
}

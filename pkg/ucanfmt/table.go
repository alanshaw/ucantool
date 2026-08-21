package ucanfmt

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// newTable creates the two-column layout shared by all formatters in this
// package: ASCII borders, a line between every row and no wrapping, so
// multi-line values such as hex dumps and DAG-JSON stay intact.
func newTable(w io.Writer, header ...string) *tablewriter.Table {
	table := tablewriter.NewTable(w,
		tablewriter.WithRendition(tw.Rendition{
			Symbols:  tw.NewSymbols(tw.StyleASCII),
			Settings: tw.Settings{Separators: tw.Separators{BetweenRows: tw.On}},
		}),
		tablewriter.WithRowAutoWrap(tw.WrapNone),
		tablewriter.WithRowAlignment(tw.AlignLeft),
	)
	table.Header(header)
	return table
}

func renderTable(table *tablewriter.Table) {
	if err := table.Render(); err != nil {
		panic(fmt.Errorf("rendering table: %w", err))
	}
}

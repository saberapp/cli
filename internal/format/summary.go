package format

import (
	"fmt"
	"io"

	"github.com/saberapp/cli/internal/client"
)

// PrintSummary renders a generated summary.
func PrintSummary(w io.Writer, summary []client.DataPoint) {
	if len(summary) == 0 {
		fmt.Fprintln(w, "No data points in summary.")
		return
	}
	for i, dp := range summary {
		fmt.Fprintf(w, "%d. %s\n", i+1, dp.Description)
		fmt.Fprintf(w, "   Qualification: %s\n", dp.Qualification)
		if len(dp.ReferenceQuestions) > 0 {
			fmt.Fprintf(w, "   Based on: %s\n", JoinStrings(dp.ReferenceQuestions, "\u2014"))
		}
		if len(dp.Sources) > 0 {
			max := len(dp.Sources)
			if max > 2 {
				max = 2
			}
			for _, s := range dp.Sources[:max] {
				fmt.Fprintf(w, "   Source: %s (%s)\n", s.Title, s.URL)
			}
		}
		if i < len(summary)-1 {
			fmt.Fprintln(w)
		}
	}
}

// PrintSummaryRecords renders a table of summary records.
func PrintSummaryRecords(w io.Writer, records []client.SummaryRecord, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tSTATUS\tSIGNALS\tDATA POINTS\tCREATED")
	for _, r := range records {
		dpCount := len(r.Summary)
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n",
			r.ID,
			r.Status,
			r.SignalsCount,
			dpCount,
			r.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d summaries\n", len(records), total)
}

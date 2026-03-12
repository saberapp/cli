package format

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// NewTabWriter returns a tabwriter suitable for CLI tables.
func NewTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

// FlushTable flushes a tabwriter, discarding the write error.
// Writes to stdout/stderr in a CLI tool cannot meaningfully fail.
func FlushTable(tw *tabwriter.Writer) {
	_ = tw.Flush()
}

// KV prints key-value rows using a tabwriter.
func KV(w io.Writer, rows [][2]string) {
	tw := NewTabWriter(w)
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	FlushTable(tw)
}

// JoinStrings joins a string slice with commas, or returns fallback if empty.
func JoinStrings(ss []string, fallback string) string {
	if len(ss) == 0 {
		return fallback
	}
	return strings.Join(ss, ", ")
}

// TruncateString truncates s to n chars, appending "..." if truncated.
func TruncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

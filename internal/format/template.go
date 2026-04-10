package format

import (
	"fmt"
	"io"

	"github.com/saberapp/cli/internal/client"
)

// PrintSignalTemplate renders a single signal template.
func PrintSignalTemplate(w io.Writer, tmpl *client.SignalTemplate) {
	weight := tmpl.Weight
	if weight == "" {
		weight = "\u2014"
	}
	description := tmpl.Description
	if description == "" {
		description = "\u2014"
	}
	deletedAt := "\u2014"
	if tmpl.DeletedAt != nil {
		deletedAt = tmpl.DeletedAt.UTC().Format("2006-01-02 15:04 UTC")
	}
	KV(w, [][2]string{
		{"ID:", tmpl.ID},
		{"Name:", tmpl.Name},
		{"Version:", fmt.Sprintf("%d", tmpl.Version)},
		{"Question:", tmpl.Question},
		{"Answer type:", tmpl.AnswerType},
		{"Weight:", weight},
		{"Description:", TruncateString(description, 100)},
		{"Source:", tmpl.Source},
		{"Created:", tmpl.CreatedAt.UTC().Format("2006-01-02 15:04 UTC")},
		{"Deleted:", deletedAt},
	})
}

// PrintSignalTemplates renders a table of signal templates.
func PrintSignalTemplates(w io.Writer, templates []client.SignalTemplate, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tNAME\tQUESTION\tANSWER TYPE\tVERSION")
	for _, t := range templates {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
			t.ID,
			TruncateString(t.Name, 30),
			TruncateString(t.Question, 40),
			t.AnswerType,
			t.Version,
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d templates\n", len(templates), total)
}

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

// PrintExtractProposal renders the result of a `template extract propose` call.
// Splits clusters by kind (existing-template attaches vs new-template proposals)
// so the user can see at a glance what the LLM decided.
func PrintExtractProposal(w io.Writer, p *client.ExtractProposal) {
	if len(p.Clusters) == 0 {
		fmt.Fprintf(w, "No clusters proposed (%d/%d candidates processed).\n",
			p.ProcessedCandidates, p.TotalCandidates)
		if p.HasMore {
			fmt.Fprintln(w, "More candidates remain — re-run propose to process the next page.")
		}
		return
	}

	newCount, existingCount := 0, 0
	for _, c := range p.Clusters {
		if c.Kind == "existing" {
			existingCount++
		} else {
			newCount++
		}
	}
	totalExec := 0
	for _, c := range p.Clusters {
		totalExec += len(c.ExecutionIDs)
	}

	fmt.Fprintf(w, "%d clusters from %d/%d candidates (%d new templates, %d attaches to existing, %d executions)\n\n",
		len(p.Clusters), p.ProcessedCandidates, p.TotalCandidates,
		newCount, existingCount, totalExec)

	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "KIND\tNAME / TEMPLATE\tANSWER TYPE\tEXECS\tQUESTION / SAMPLE")
	for _, c := range p.Clusters {
		nameOrTpl := c.Name
		question := TruncateString(c.Question, 50)
		if c.Kind == "existing" {
			nameOrTpl = c.TemplateID
			if len(c.SampleQuestions) > 0 {
				question = TruncateString(c.SampleQuestions[0], 50)
			}
		}
		answerType := c.AnswerType
		if answerType == "" {
			answerType = "—"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n",
			c.Kind,
			TruncateString(nameOrTpl, 36),
			answerType,
			len(c.ExecutionIDs),
			question,
		)
	}
	FlushTable(tw)

	if p.HasMore {
		fmt.Fprintln(w, "\nMore candidates remain — re-run propose after applying to process the next page.")
	}
}

// PrintExtractApplyResult renders the result of a `template extract apply` call.
func PrintExtractApplyResult(w io.Writer, r *client.ExtractApplyResult) {
	if len(r.Created) == 0 {
		fmt.Fprintln(w, "No clusters applied.")
		return
	}

	newCount, existingCount, totalAttached := 0, 0, 0
	for _, c := range r.Created {
		if c.Kind == "existing" {
			existingCount++
		} else {
			newCount++
		}
		totalAttached += c.Attached
	}

	fmt.Fprintf(w, "Applied %d clusters: %d new templates, %d attaches to existing, %d executions attached\n\n",
		len(r.Created), newCount, existingCount, totalAttached)

	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "KIND\tTEMPLATE ID\tNAME\tATTACHED")
	for _, c := range r.Created {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\n",
			c.Kind,
			c.TemplateID,
			TruncateString(c.Name, 40),
			c.Attached,
		)
	}
	FlushTable(tw)
}

package format

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/saberapp/cli/internal/client"
)

// PrintScoringProfile renders a single scoring profile.
func PrintScoringProfile(w io.Writer, p *client.ScoringProfile) {
	desc := "—"
	if p.Description != nil && *p.Description != "" {
		desc = *p.Description
	}
	KV(w, [][2]string{
		{"ID:", p.ID},
		{"Name:", p.Name},
		{"Type:", p.Type},
		{"Description:", desc},
		{"Created:", p.CreatedAt.Format("2006-01-02 15:04:05")},
		{"Updated:", p.UpdatedAt.Format("2006-01-02 15:04:05")},
	})
}

// PrintScoringProfiles renders a table of scoring profiles.
func PrintScoringProfiles(w io.Writer, profiles []client.ScoringProfile) {
	if len(profiles) == 0 {
		fmt.Fprintln(w, "No scoring profiles found.")
		return
	}
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tNAME\tTYPE\tCREATED")
	for _, p := range profiles {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			p.ID,
			TruncateString(p.Name, 40),
			p.Type,
			p.CreatedAt.Format("2006-01-02"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d profiles\n", len(profiles))
}

// PrintScoringRule renders a single scoring rule with its point values.
func PrintScoringRule(w io.Writer, r *client.ScoringRule) {
	KV(w, [][2]string{
		{"ID:", r.ID},
		{"Profile:", r.ProfileID},
		{"Signal template:", r.SignalTemplateID},
		{"Dimension:", r.Dimension},
		{"Point values:", summarisePointValues(r.PointValues)},
		{"Created:", r.CreatedAt.Format("2006-01-02 15:04:05")},
	})
}

// PrintScoringRules renders a table of scoring rules.
func PrintScoringRules(w io.Writer, rules []client.ScoringRule) {
	if len(rules) == 0 {
		fmt.Fprintln(w, "No scoring rules in this profile.")
		return
	}
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tDIMENSION\tSIGNAL TEMPLATE\tPOINT VALUES")
	for _, r := range rules {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			r.ID,
			r.Dimension,
			r.SignalTemplateID,
			TruncateString(summarisePointValues(r.PointValues), 60),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d rules\n", len(rules))
}

// PrintProfileAssignment renders a single profile assignment.
func PrintProfileAssignment(w io.Writer, a *client.ProfileAssignment) {
	KV(w, [][2]string{
		{"ID:", a.ID},
		{"Profile:", a.ProfileID},
		{"Object type:", a.ObjectType},
		{"Object ID:", a.ObjectID},
		{"Assigned:", a.AssignedAt.Format("2006-01-02 15:04:05")},
	})
}

// PrintProfileAssignments renders a table of profile assignments.
func PrintProfileAssignments(w io.Writer, assignments []client.ProfileAssignment) {
	if len(assignments) == 0 {
		fmt.Fprintln(w, "No profile assignments found.")
		return
	}
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tPROFILE\tOBJECT TYPE\tOBJECT ID\tASSIGNED")
	for _, a := range assignments {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			a.ID,
			a.ProfileID,
			a.ObjectType,
			TruncateString(a.ObjectID, 40),
			a.AssignedAt.Format("2006-01-02"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d assignments\n", len(assignments))
}

// PrintScoreResult renders a single score with its contributions.
func PrintScoreResult(w io.Writer, s *client.ScoreResult) {
	delta := "—"
	if s.PreviousScore != nil {
		delta = fmt.Sprintf("%s (was %s)", trimFloat(s.Score), trimFloat(*s.PreviousScore))
	} else {
		delta = trimFloat(s.Score)
	}
	KV(w, [][2]string{
		{"Profile:", s.ProfileID},
		{"Object:", fmt.Sprintf("%s %s", s.ObjectType, s.ObjectID)},
		{"Dimension:", s.Dimension},
		{"Score:", delta},
		{"Coverage:", fmt.Sprintf("%d / %d rules", s.SignalCoverage, s.TotalRules)},
		{"Computed:", s.ComputedAt.Format("2006-01-02 15:04:05")},
		{"Version:", strconv.Itoa(s.Version)},
	})
	if len(s.Contributions) == 0 {
		return
	}
	fmt.Fprintln(w, "\nContributions:")
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "  SIGNAL TEMPLATE\tMATCHED\tEARNED\tMAX")
	for _, c := range s.Contributions {
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n",
			c.SignalTemplateID,
			TruncateString(c.MatchedValue, 30),
			trimFloat(c.PointsEarned),
			trimFloat(c.MaxPoints),
		)
	}
	FlushTable(tw)
}

// PrintScoreResults renders a compact table of scores across objects/dimensions.
func PrintScoreResults(w io.Writer, scores []client.ScoreResult) {
	if len(scores) == 0 {
		fmt.Fprintln(w, "No scores found.")
		return
	}
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "OBJECT\tPROFILE\tDIMENSION\tSCORE\tCOVERAGE\tCOMPUTED")
	for _, s := range scores {
		score := trimFloat(s.Score)
		if s.PreviousScore != nil {
			score = fmt.Sprintf("%s (was %s)", score, trimFloat(*s.PreviousScore))
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d/%d\t%s\n",
			TruncateString(s.ObjectID, 30),
			s.ProfileID,
			s.Dimension,
			score,
			s.SignalCoverage,
			s.TotalRules,
			s.ComputedAt.Format("2006-01-02 15:04"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d scores\n", len(scores))
}

// summarisePointValues produces a compact human-readable form of a rule's point values.
func summarisePointValues(p client.ScoringPointValues) string {
	switch {
	case p.True != nil || p.False != nil:
		t, f := "—", "—"
		if p.True != nil {
			t = trimFloat(*p.True)
		}
		if p.False != nil {
			f = trimFloat(*p.False)
		}
		return fmt.Sprintf("true=%s, false=%s", t, f)
	case len(p.Ranges) > 0:
		parts := make([]string, 0, len(p.Ranges))
		for _, r := range p.Ranges {
			parts = append(parts, fmt.Sprintf("%s–%s→%s",
				trimFloat(r.Min), trimFloat(r.Max), trimFloat(r.Points)))
		}
		return joinStrings(parts, ", ")
	case len(p.Choices) > 0:
		keys := make([]string, 0, len(p.Choices))
		for k := range p.Choices {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s→%s", k, trimFloat(p.Choices[k])))
		}
		return joinStrings(parts, ", ")
	default:
		return "—"
	}
}

// trimFloat formats a float without trailing zeros (e.g. 25.0 → "25", 12.5 → "12.5").
func trimFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// joinStrings concatenates with sep, used here to avoid pulling in strings just for one helper.
func joinStrings(ss []string, sep string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

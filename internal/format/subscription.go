package format

import (
	"fmt"
	"io"

	"github.com/saberapp/cli/internal/client"
)

// PrintSubscription renders a single subscription.
func PrintSubscription(w io.Writer, sub *client.Subscription) {
	schedule := sub.Frequency
	if schedule == "" {
		schedule = sub.CronExpression
	}
	lastRun := "—"
	if sub.LastRunAt != nil {
		lastRun = sub.LastRunAt.UTC().Format("2006-01-02 15:04 UTC")
	}
	nextRun := "—"
	if sub.NextRunAt != nil {
		nextRun = sub.NextRunAt.UTC().Format("2006-01-02 15:04 UTC")
	}
	KV(w, [][2]string{
		{"ID:", sub.ID},
		{"Name:", sub.Name},
		{"Question:", TruncateString(sub.Question, 80)},
		{"Answer type:", sub.AnswerType},
		{"Schedule:", schedule},
		{"Status:", sub.Status},
		{"List ID:", sub.ListID},
		{"Last run:", lastRun},
		{"Next run:", nextRun},
		{"Created:", sub.CreatedAt.UTC().Format("2006-01-02 15:04 UTC")},
	})
}

// PrintSubscriptions renders a table of subscriptions.
func PrintSubscriptions(w io.Writer, subs []client.Subscription, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tNAME\tSCHEDULE\tSTATUS\tLAST RUN")
	for _, s := range subs {
		schedule := s.Frequency
		if schedule == "" {
			schedule = TruncateString(s.CronExpression, 15)
		}
		lastRun := "—"
		if s.LastRunAt != nil {
			lastRun = s.LastRunAt.UTC().Format("2006-01-02")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			s.ID,
			TruncateString(s.Name, 35),
			schedule,
			s.Status,
			lastRun,
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d subscriptions\n", len(subs), total)
}

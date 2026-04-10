package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/saberapp/cli/internal/client"
)

// PrintMarketSignalSubscription renders a single market signal subscription.
func PrintMarketSignalSubscription(w io.Writer, sub *client.MarketSignalSubscription) {
	name := "\u2014"
	if sub.Name != nil {
		name = *sub.Name
	}
	prompt := "\u2014"
	if sub.Prompt != nil && *sub.Prompt != "" {
		prompt = TruncateString(*sub.Prompt, 120)
	}

	rows := [][2]string{
		{"ID:", sub.ID},
		{"Type:", sub.Type},
		{"Name:", name},
		{"Status:", sub.Status},
		{"Interval:", sub.Interval},
		{"Signal limit:", fmt.Sprintf("%d", sub.IntervalSignalLimit)},
		{"Webhook:", sub.WebhookURL},
		{"Prompt:", prompt},
		{"Created:", sub.CreatedAt.UTC().Format("2006-01-02 15:04 UTC")},
		{"Updated:", sub.UpdatedAt.UTC().Format("2006-01-02 15:04 UTC")},
	}
	KV(w, rows)

	if len(sub.Filters) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Filters:")
		printFilters(w, sub.Filters)
	}
}

// printFilters prints filter key-value pairs with indentation.
func printFilters(w io.Writer, filters map[string]any) {
	for k, v := range filters {
		switch val := v.(type) {
		case []any:
			parts := make([]string, len(val))
			for i, item := range val {
				parts[i] = fmt.Sprintf("%v", item)
			}
			fmt.Fprintf(w, "  %s: %s\n", k, strings.Join(parts, ", "))
		default:
			fmt.Fprintf(w, "  %s: %v\n", k, v)
		}
	}
}

// PrintMarketSignalSubscriptions renders a table of subscriptions.
func PrintMarketSignalSubscriptions(w io.Writer, subs []client.MarketSignalSubscription, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tTYPE\tNAME\tSTATUS\tINTERVAL\tCREATED")
	for _, s := range subs {
		name := "\u2014"
		if s.Name != nil {
			name = TruncateString(*s.Name, 30)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.ID,
			s.Type,
			name,
			s.Status,
			s.Interval,
			s.CreatedAt.UTC().Format("2006-01-02"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d subscriptions\n", len(subs), total)
}

// PrintMarketSignals renders a table of market signals.
func PrintMarketSignals(w io.Writer, signals []client.MarketSignal, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tSTATUS\tCONFIDENCE\tPUBLISHED\tCREATED")
	for _, s := range signals {
		confidence := "\u2014"
		if s.ConfidenceScore != nil {
			confidence = fmt.Sprintf("%.0f%%", *s.ConfidenceScore*100)
		}
		published := "\u2014"
		if s.PublishedAt != nil {
			published = s.PublishedAt.UTC().Format("2006-01-02 15:04")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			s.ID,
			s.Status,
			confidence,
			published,
			s.CreatedAt.UTC().Format("2006-01-02 15:04"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d signals\n", len(signals), total)
}

// PrintMarketSignal renders a single market signal detail.
func PrintMarketSignal(w io.Writer, sig *client.MarketSignal) {
	confidence := "\u2014"
	if sig.ConfidenceScore != nil {
		confidence = fmt.Sprintf("%.0f%%", *sig.ConfidenceScore*100)
	}
	published := "\u2014"
	if sig.PublishedAt != nil {
		published = sig.PublishedAt.UTC().Format("2006-01-02 15:04 UTC")
	}
	delivered := "\u2014"
	if sig.DeliveredAt != nil {
		delivered = sig.DeliveredAt.UTC().Format("2006-01-02 15:04 UTC")
	}

	KV(w, [][2]string{
		{"ID:", sig.ID},
		{"Subscription:", sig.SubscriptionID},
		{"Status:", sig.Status},
		{"Confidence:", confidence},
		{"Published:", published},
		{"Delivered:", delivered},
		{"Created:", sig.CreatedAt.UTC().Format("2006-01-02 15:04 UTC")},
	})

	if len(sig.Payload) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Payload:")
		b, err := json.MarshalIndent(sig.Payload, "  ", "  ")
		if err == nil {
			fmt.Fprintf(w, "  %s\n", string(b))
		}
	}
}

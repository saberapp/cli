package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/saberapp/cli/internal/client"
)

// PrintSignal renders a signal result to w.
func PrintSignal(w io.Writer, sig *client.Signal) {
	subject := [2]string{"Domain:", sig.Domain}
	if sig.ContactProfileURL != "" {
		subject = [2]string{"Profile:", sig.ContactProfileURL}
	}
	rows := [][2]string{
		subject,
		{"Question:", sig.Question},
	}

	if sig.Status == client.SignalStatusFailed {
		rows = append(rows, [2]string{"Status:", "failed"})
		if sig.Error != "" {
			rows = append(rows, [2]string{"Error:", sig.Error})
		}
		KV(w, rows)
		return
	}

	rows = append(rows, [2]string{"Answer:", formatAnswer(sig.Answer)})

	if sig.Confidence != nil {
		rows = append(rows, [2]string{"Confidence:", fmt.Sprintf("%.0f%%", *sig.Confidence*100)})
	}
	if sig.Reasoning != "" {
		rows = append(rows, [2]string{"Reasoning:", TruncateString(sig.Reasoning, 200)})
	}

	KV(w, rows)

	if len(sig.Sources) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Sources:")
		max := len(sig.Sources)
		if max > 3 {
			max = 3
		}
		for i, s := range sig.Sources[:max] {
			fmt.Fprintf(w, "  %d. %s\n     %s\n", i+1, s.Title, s.URL)
		}
	}
}

// PrintSignalCreated renders a minimal "signal created" message (for --no-wait).
func PrintSignalCreated(w io.Writer, sig *client.Signal) {
	fmt.Fprintf(w, "Signal created  ID: %s  Status: %s\n", sig.ID, sig.Status)
}

// PrintSignalList renders a table of signal list items.
func PrintSignalList(w io.Writer, signals []client.SignalListItem, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tDOMAIN\tQUESTION\tSTATUS\tANSWER TYPE\tCREATED")
	for _, s := range signals {
		domain := s.Domain
		if domain == "" && s.ContactProfileURL != "" {
			domain = TruncateString(s.ContactProfileURL, 30)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.ID,
			domain,
			TruncateString(s.Question, 35),
			string(s.Status),
			s.AnswerType,
			s.CreatedAt.UTC().Format("2006-01-02 15:04"),
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d signals\n", len(signals), total)
}

// PrintBatchResult renders the result of a batch signal creation.
func PrintBatchResult(w io.Writer, resp *client.SignalBatchResponse) {
	if resp.IsAsync {
		KV(w, [][2]string{
			{"Mode:", "async"},
			{"Batch ID:", resp.BatchID},
			{"Total signals:", fmt.Sprintf("%d", resp.TotalSignals)},
			{"Status:", resp.Status},
			{"Submitted:", resp.SubmittedAt},
		})
		return
	}
	KV(w, [][2]string{
		{"Total signals:", fmt.Sprintf("%d", resp.TotalSignals)},
		{"Accepted:", fmt.Sprintf("%d", resp.Accepted)},
		{"Rejected:", fmt.Sprintf("%d", resp.Rejected)},
		{"Submitted:", resp.SubmittedAt},
	})
	if len(resp.Results) > 0 {
		fmt.Fprintln(w)
		tw := NewTabWriter(w)
		fmt.Fprintln(tw, "ID\tDOMAIN\tQUESTION\tSTATUS")
		for _, r := range resp.Results {
			id := r.ID
			if id == "" {
				id = "\u2014"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				id,
				r.Domain,
				TruncateString(r.Question, 40),
				r.Status,
			)
		}
		FlushTable(tw)
	}
}

// formatAnswer extracts the human-readable value from the API's typed answer envelope.
// The API returns: {"type":"open_text","openText":{"value":"Yes"}}
//
//	{"type":"number","number":{"value":14491}}
//	{"type":"boolean","boolean":{"value":true}}
//	{"type":"list","list":{"value":["item1","item2"]}}  etc.
func formatAnswer(answer any) string {
	if answer == nil {
		return "—"
	}

	// Marshal back to JSON so we can decode with known keys.
	b, err := json.Marshal(answer)
	if err != nil {
		return fmt.Sprintf("%v", answer)
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(b, &envelope); err != nil {
		// Not an object — scalar value.
		return string(b)
	}

	// The envelope has a "type" key and a key matching the type name.
	var answerType string
	if raw, ok := envelope["type"]; ok {
		_ = json.Unmarshal(raw, &answerType)
	}

	// Try to extract the value from the typed sub-object.
	// Possible keys: openText, number, boolean, list, percentage, currency, url
	typeKeys := []string{answerType, "openText", "number", "boolean", "list", "percentage", "currency", "url"}
	for _, key := range typeKeys {
		raw, ok := envelope[key]
		if !ok {
			continue
		}
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(raw, &sub); err != nil {
			continue
		}
		valRaw, ok := sub["value"]
		if !ok {
			continue
		}
		// Try array first (list answer type).
		var arr []any
		if err := json.Unmarshal(valRaw, &arr); err == nil {
			parts := make([]string, len(arr))
			for i, item := range arr {
				parts[i] = fmt.Sprintf("%v", item)
			}
			return strings.Join(parts, ", ")
		}
		// Scalar value.
		var v any
		if err := json.Unmarshal(valRaw, &v); err == nil {
			return fmt.Sprintf("%v", v)
		}
	}

	// Fallback: compact JSON.
	return string(b)
}

package format

import (
	"bytes"
	"math"
	"strconv"
	"testing"

	"github.com/saberapp/cli/internal/client"
)

func TestPrintPhoneRun(t *testing.T) {
	amount := math.Pi
	zero := 0.0
	for _, tt := range []struct {
		name string
		run  client.PhoneRun
		want string
	}{
		{
			name: "found display",
			run:  client.PhoneRun{Status: "completed", Result: &client.PhoneResult{Found: true, Phone: &client.Phone{E164: "+447700900123", Display: "+44 7700 900123", LineType: "mobile", Country: "GB"}}, Credits: &client.RunCredits{State: "charged", Amount: &amount}},
			want: "Phone: +44 7700 900123\nLine type: mobile\nCountry: GB\nRun credits: charged " + strconv.FormatFloat(amount, 'f', -1, 64) + "\n",
		},
		{
			name: "found e164 fallback no credits",
			run:  client.PhoneRun{Status: "completed", Result: &client.PhoneResult{Found: true, Phone: &client.Phone{E164: "+447700900123", LineType: "mobile", Country: "GB"}}},
			want: "Phone: +447700900123\nLine type: mobile\nCountry: GB\n",
		},
		{
			name: "not found",
			run:  client.PhoneRun{Status: "completed", Result: &client.PhoneResult{Found: false}, Credits: &client.RunCredits{State: "released"}},
			want: "No mobile number found.\nRun credits: released\n",
		},
		{
			name: "processing",
			run:  client.PhoneRun{ID: "example-run", Status: "processing", Credits: &client.RunCredits{State: "reserved"}},
			want: "Phone lookup still processing  ID: example-run\nRe-run later with: saber contact find-phone get example-run\n",
		},
		{
			name: "failed retryable",
			run:  client.PhoneRun{Status: "failed", Error: "lookup unavailable", ErrorCode: "PHONE_LOOKUP_UNAVAILABLE", Retryable: true, Credits: &client.RunCredits{State: "released"}},
			want: "Phone lookup failed: lookup unavailable (PHONE_LOOKUP_UNAVAILABLE)\nYou can retry.\n",
		},
		{
			name: "failed permanent",
			run:  client.PhoneRun{Status: "failed", Error: "lookup rejected", ErrorCode: "PHONE_LOOKUP_REJECTED"},
			want: "Phone lookup failed: lookup rejected (PHONE_LOOKUP_REJECTED)\n",
		},
		{
			name: "unknown credit state and zero amount",
			run:  client.PhoneRun{Status: "completed", Result: &client.PhoneResult{Found: false}, Credits: &client.RunCredits{State: "future-state", Amount: &zero}},
			want: "No mobile number found.\nRun credits: future-state 0\n",
		},
		{
			name: "unknown status",
			run:  client.PhoneRun{ID: "example-run", Status: "future-status"},
			want: "Phone lookup status: future-status  ID: example-run\n",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			PrintPhoneRun(&out, &tt.run)
			if out.String() != tt.want {
				t.Fatalf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

package format

import (
	"fmt"
	"io"
	"strconv"

	"github.com/saberapp/cli/internal/client"
)

func PrintPhoneRun(w io.Writer, run *client.PhoneRun) {
	switch run.Status {
	case "processing":
		fmt.Fprintf(w, "Phone lookup still processing  ID: %s\n", run.ID)
		fmt.Fprintf(w, "Re-run later with: saber contact find-phone get %s\n", run.ID)
	case "completed":
		if run.Result != nil && run.Result.Found && run.Result.Phone != nil {
			phone := run.Result.Phone
			number := phone.Display
			if number == "" {
				number = phone.E164
			}
			fmt.Fprintf(w, "Phone: %s\nLine type: %s\nCountry: %s\n", number, phone.LineType, phone.Country)
		} else {
			fmt.Fprintln(w, "No mobile number found.")
		}
	case "failed":
		fmt.Fprintf(w, "Phone lookup failed: %s (%s)\n", run.Error, run.ErrorCode)
		if run.Retryable {
			fmt.Fprintln(w, "You can retry.")
		}
	default:
		fmt.Fprintf(w, "Phone lookup status: %s  ID: %s\n", run.Status, run.ID)
	}
	if run.Status == "completed" && run.Credits != nil {
		fmt.Fprintf(w, "Run credits: %s", run.Credits.State)
		if run.Credits.Amount != nil {
			fmt.Fprintf(w, " %s", strconv.FormatFloat(*run.Credits.Amount, 'f', -1, 64))
		}
		fmt.Fprintln(w)
	}
}

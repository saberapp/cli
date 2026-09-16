package telemetry

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// PromptLines is the first-run consent question, printed to stderr so stdout
// stays clean for piped output. It states what is sent, what is never sent,
// and how to change the answer later.
var PromptLines = []string{
	"Saber CLI can send anonymous usage data: the command name, which flag names you used,",
	"the CLI version, your OS, and how long the command took.",
	"It never sends your API key, your arguments, flag values, or any result data.",
	"This helps us see which commands need work. Recommended: yes.",
	"You can change this at any time with \"saber telemetry disable\".",
}

// PromptQuestion is the single question line. Enter accepts the default (yes).
const PromptQuestion = "Send anonymous usage data? [Y/n] "

// Ask prints the consent question to out and reads one line from in.
// Enter, "y", or "yes" mean yes. Anything else means no. A read failure means
// no, so an unreadable terminal never turns telemetry on by accident.
func Ask(in io.Reader, out io.Writer) bool {
	for _, line := range PromptLines {
		fmt.Fprintln(out, line)
	}
	fmt.Fprint(out, PromptQuestion)

	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && answer == "" {
		fmt.Fprintln(out)
		return false
	}
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "", "y", "yes":
		fmt.Fprintln(out, "Usage data is on. Turn it off with \"saber telemetry disable\".")
		return true
	default:
		fmt.Fprintln(out, "Usage data is off.")
		return false
	}
}

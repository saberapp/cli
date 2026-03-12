package format

import (
	"io"
	"time"

	"github.com/briandowns/spinner"
	"golang.org/x/term"
)

// Spinner wraps briandowns/spinner with TTY detection.
type Spinner struct {
	s       *spinner.Spinner
	enabled bool
}

// NewSpinner creates a spinner that only renders when stdout is a TTY
// and quiet/json modes are off.
func NewSpinner(w io.Writer, suppress bool) *Spinner {
	if suppress {
		return &Spinner{enabled: false}
	}
	// Check if w is a TTY; cast to check fd.
	isTTY := false
	type fder interface{ Fd() uintptr }
	if f, ok := w.(fder); ok {
		isTTY = term.IsTerminal(int(f.Fd()))
	}
	if !isTTY {
		return &Spinner{enabled: false}
	}
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(w))
	return &Spinner{s: s, enabled: true}
}

// Start starts the spinner with the given suffix text.
func (sp *Spinner) Start(msg string) {
	if !sp.enabled {
		return
	}
	sp.s.Suffix = " " + msg
	sp.s.Start()
}

// Stop stops the spinner and clears it.
func (sp *Spinner) Stop() {
	if !sp.enabled || sp.s == nil {
		return
	}
	sp.s.Stop()
}

// Update changes the spinner message while running.
func (sp *Spinner) Update(msg string) {
	if !sp.enabled || sp.s == nil {
		return
	}
	sp.s.Suffix = " " + msg
}

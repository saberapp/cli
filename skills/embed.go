// Package skills provides embedded Saber skill files for Claude Code.
package skills

import _ "embed"

//go:embed saber-signal-discovery.md
var SignalDiscovery string

//go:embed saber-create-company-signals.md
var CreateCompanySignals string

//go:embed saber-create-contact-signals.md
var CreateContactSignals string

//go:embed saber-build-account-list.md
var BuildAccountList string

//go:embed saber-build-contact-list.md
var BuildContactList string

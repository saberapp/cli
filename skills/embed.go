// Package skills provides embedded Saber skill files for agent skill installation.
package skills

import "embed"

// SaberCLI contains the complete saber-cli skill directory tree
// (SKILL.md + references/*). Use fs.ReadFile or fs.ReadDir to access files.
//
//go:embed saber-cli/SKILL.md saber-cli/references/*.md
var SaberCLI embed.FS

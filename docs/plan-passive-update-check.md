# Passive Update Check - Implementation Plan

## Goal

Add a non-blocking, throttled background version check so users see a concise
"update available" notice on regular command runs, without slowing down any
command or spamming the message too often.

---

## Design Decisions

| Decision               | Choice                                                                 | Rationale                                                                                                                                                 |
| ---------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Check frequency        | Once every **24 hours**                                                | Industry standard (gh, npm, terraform, flyctl all use 24h). Avoids GitHub API rate limits and keeps things quiet.                                         |
| Blocking vs background | **Background goroutine**                                               | The HTTP call runs in a goroutine; the command never waits for it. If the check finishes before the command exits, we print. If not, nothing happens.     |
| Cache location         | `~/.saber/update-check.json`                                           | Already use `~/.saber/` for credentials. Keeps everything in one place.                                                                                   |
| Output target          | **stderr**                                                             | Matches convention (gh, npm). Keeps stdout clean for `--json` and piped output.                                                                           |
| Message format         | Single line, no box art                                                | AI agents and scripts need minimal noise. One line like: `Notice: A new version of saber is available (v0.1.5 -> v0.1.7). Run "saber update" to upgrade.` |
| Opt-out                | `SABER_NO_UPDATE_CHECK=1` env var                                      | Standard escape hatch for CI, scripts, and users who do not want it.                                                                                      |
| Skip conditions        | `--json`, `--quiet`, dev builds, non-TTY stderr, `saber update` itself | Prevents noise in machine-readable output and avoids double-checking during the explicit update command.                                                  |

---

## Files to Create / Modify

### 1. `internal/update/update.go` (NEW)

Core logic for the background update checker.

```
Package: update

Types:
  - State struct       -- JSON-serializable cache { LatestVersion, CheckedAt }

Functions:
  - CacheFilePath() string
      Returns ~/.saber/update-check.json

  - LoadState() (*State, error)
      Reads and unmarshals the cache file. Returns zero-value State on
      missing/corrupt file (never errors fatally).

  - SaveState(state *State) error
      Writes cache atomically (write to .tmp, rename).

  - ShouldCheck(state *State, interval time.Duration) bool
      Returns true if CheckedAt is zero or older than interval.

  - FetchLatestVersion() (string, error)
      Moved/refactored from cmd/update.go:fetchLatestCLIVersion().
      Shared between the explicit `saber update` command and the background
      checker.

  - RunBackgroundCheck(currentVersion string, interval time.Duration) <-chan string
      1. Loads state; if !ShouldCheck -> returns closed channel (no message).
      2. Spawns goroutine: fetches latest, saves state, sends update message
         string into channel if newer version found, then closes channel.
      3. Caller can select on channel or ignore it.

  - FormatNotice(current, latest string) string
      Returns the single-line notice string.
```

### 2. `internal/update/update_test.go` (NEW)

Unit tests:

- `TestShouldCheck` -- fresh state, stale state, recent state
- `TestLoadState` -- missing file, corrupt file, valid file
- `TestSaveState` -- writes and reads back
- `TestFormatNotice` -- output format
- `TestRunBackgroundCheck` -- with a mock HTTP server returning a fake release

### 3. `cmd/update.go` (MODIFY)

- Remove `fetchLatestCLIVersion()` from this file.
- Import `internal/update` and call `update.FetchLatestVersion()` instead.
- Everything else (brew upgrade logic, `isHomebrewInstall`) stays.

### 4. `cmd/root.go` (MODIFY)

Add a `PersistentPreRun` hook and a `PersistentPostRun` hook:

```go
var updateCh <-chan string  // package-level, set in PersistentPreRun

PersistentPreRun: func(cmd *cobra.Command, args []string) {
    // Skip conditions:
    // - version == "dev"
    // - --json or --quiet flag set
    // - SABER_NO_UPDATE_CHECK=1
    // - command is "update" (already checks)
    // - stderr is not a TTY
    if shouldSkipUpdateCheck(cmd) {
        return
    }
    updateCh = update.RunBackgroundCheck(cliVersion, 24*time.Hour)
}

PersistentPostRun: func(cmd *cobra.Command, args []string) {
    if updateCh == nil {
        return
    }
    // Non-blocking read: if the goroutine finished, print; otherwise skip.
    select {
    case msg, ok := <-updateCh:
        if ok && msg != "" {
            fmt.Fprintln(os.Stderr, msg)
        }
    default:
        // Check did not finish in time, silently skip.
    }
}
```

Helper:

```go
func shouldSkipUpdateCheck(cmd *cobra.Command) bool {
    if cliVersion == "dev" { return true }
    if jsonOutput || quiet { return true }
    if os.Getenv("SABER_NO_UPDATE_CHECK") != "" { return true }
    if cmd.Name() == "update" { return true }
    if !term.IsTerminal(int(os.Stderr.Fd())) { return true }
    return false
}
```

### 5. `internal/config/config.go` (MODIFY - minor)

Export `SaberDir()` (capitalize) so the update package can use it to find
`~/.saber/`. Alternatively, duplicate the 3-line helper in the update package
to avoid coupling.

---

## Message Format

```
Notice: Update available v0.1.5 -> v0.1.7. Run "saber update" to upgrade.
```

Single line. No color codes (keeps it safe for all terminals and log parsers).
Dim ANSI is acceptable if stderr is a TTY, but not required.

---

## Edge Cases

| Case                                                  | Behavior                                                       |
| ----------------------------------------------------- | -------------------------------------------------------------- |
| No internet / GitHub down                             | Goroutine fails silently; no message printed. Cache unchanged. |
| Corrupt cache file                                    | Treated as "never checked"; triggers a fresh check.            |
| `~/.saber/` does not exist yet                        | `SaveState` creates it (MkdirAll).                             |
| Race between two concurrent saber processes           | Both may check; last writer wins on the cache file. Harmless.  |
| Version downgrade (user on newer than latest release) | No message printed (only prints when latest > current).        |

---

## Implementation Order

1. Create `internal/update/update.go` with all the logic.
2. Create `internal/update/update_test.go` with tests.
3. Refactor `cmd/update.go` to use `update.FetchLatestVersion()`.
4. Wire up `PersistentPreRun`/`PersistentPostRun` in `cmd/root.go`.
5. Add `shouldSkipUpdateCheck` helper in `cmd/root.go`.
6. Run `go test ./...` and `go build ./...` to verify.
7. Manual smoke test with `SABER_NO_UPDATE_CHECK`, `--json`, `--quiet`, dev build.
8. Update skill docs if needed (per AGENTS.md instructions).

---

## References

- **GitHub CLI (gh)**: 24h background check, stderr notice, `GH_NO_UPDATE_NOTIFIER` env var.
- **Terraform CLI**: 24h check, checkpoint file in config dir.
- **npm**: prints update notice after command runs, once per interval.
- **Fly.io CLI (flyctl)**: Background goroutine, non-blocking channel pattern.

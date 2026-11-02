package client

import (
	"net/http"
	"strings"
	"sync"
)

// MinVersionHeader is the response header in which the Saber API reports the
// oldest CLI version it still accepts. Older builds get HTTP 426.
const MinVersionHeader = "X-Saber-Cli-Min-Version"

// minVersion holds the last minimum version seen in this process. A CLI run is
// one command, so a process-wide value is enough; the mutex covers commands
// that call the API from more than one goroutine.
var minVersion struct {
	sync.Mutex
	value string
}

// MinSupportedVersion returns the last minimum CLI version (without a leading
// "v") that the API reported in this process, or "" when no response carried
// the header.
func MinSupportedVersion() string {
	minVersion.Lock()
	defer minVersion.Unlock()
	return minVersion.value
}

// recordMinVersion keeps the header value when it is present. A response
// without the header does not clear an earlier value.
func recordMinVersion(h http.Header) {
	v := strings.TrimPrefix(strings.TrimSpace(h.Get(MinVersionHeader)), "v")
	if v == "" {
		return
	}
	minVersion.Lock()
	minVersion.value = v
	minVersion.Unlock()
}

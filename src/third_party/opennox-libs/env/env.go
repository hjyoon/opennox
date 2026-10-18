package env

import (
	"os"
	"sync"
)

func IsDevMode() bool {
	return os.Getenv("NOX_DEV") == "true"
}

var (
	e2eOnce sync.Once
	isE2E   bool
)

func IsE2E() bool {
	e2eOnce.Do(func() {
		isE2E = os.Getenv("NOX_E2E") != "" || os.Getenv("NOX_E2E_RECORD") != ""
	})
	return isE2E
}

func AppUserDir() string {
	if os.Getenv("SNAP_NAME") == "opennox" {
		return os.Getenv("SNAP_USER_COMMON")
	}
	return ""
}

// IsE2EVirtualTime reports whether E2E playback uses the deterministic clock.
// Ordinary playback still uses the E2E seat and isolated configuration, but
// retains production tick units, time-based effects and the local calendar.
func IsE2EVirtualTime() bool {
	return IsE2E() && os.Getenv("NOX_E2E_CLOCK") != "ordinary"
}

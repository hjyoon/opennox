package legacy

import "github.com/opennox/opennox/v1/server"

// GolemDeathActualDefaultC549E90 exercises a public C entry without reference
// runtime/state overrides. Its boolean is false for an unported public ABI.
func GolemDeathActualDefaultC549E90(mech bool, unit *server.Object) (int32, bool) {
	if golemOriginalABI549E90(mech) != 64 || golemOriginalRuntimeOverride549E90 != nil || golemOriginalStateOverride549E90 != nil {
		return 0, false
	}
	return golemOriginalC549E90(mech, unit), true
}

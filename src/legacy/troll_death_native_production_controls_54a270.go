package legacy

import "github.com/opennox/opennox/v1/server"

// TrollDeathActualDefaultC54A270 calls the public C symbol with the real
// default world adapters. Reference-runtime overrides cannot earn this gate.
func TrollDeathActualDefaultC54A270(unit *server.Object) (int32, bool) {
	if trollDeathOriginalABI54A270() != 64 || trollDeathOriginalRuntimeOverride54A270 != nil || trollDeathOriginalStateOverride54A270 != nil {
		return 0, false
	}
	return trollDeathOriginalC54A270(unit), true
}

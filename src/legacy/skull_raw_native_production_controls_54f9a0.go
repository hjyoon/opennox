package legacy

import "github.com/opennox/opennox/v1/server"

// SkullRawActualDefaultC54F9A0 enters the actual public C ABI with stock
// world services. Entries are update, projectile, wrapper, scan, candidate.
func SkullRawActualDefaultC54F9A0(entry int, source, target *server.Object, id uint32) (uint32, bool) {
	if entry < 0 || entry > 4 || skullOriginalABI54F9A0(entry) != 64 || skullRawOriginalStateOverride54F9A0 != nil || skullRawOriginalRuntimeOverride54F9A0 != nil {
		return 0, false
	}
	return skullOriginalC54F9A0(entry, source, target, id), true
}

package opennox

import "github.com/opennox/opennox/v1/common/flags"

// 0050F0B3 tests 0x800 (campaign/Coop), not 0x2000 (Online).
// An ordinary arena trade does not acquire the campaign movement lock.
func shopMovementFrozen50EF10(flags noxflags.GameFlag) bool {
	return flags&noxflags.GameModeCoop != 0
}

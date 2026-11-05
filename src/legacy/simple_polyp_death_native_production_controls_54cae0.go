package legacy

import "github.com/opennox/opennox/v1/server"

// SimplePolypDeathActualDefaultC54CAE0 calls an actual public C entry with
// stock world services. Entries are ImpEgg=0, Polyp=1 and Potion=2. False
// indicates an unported ABI or an active reference runtime/state override.
func SimplePolypDeathActualDefaultC54CAE0(entry int, source *server.Object) (uint32, bool) {
	if entry < 0 || entry > 2 || simplePolypOriginalABI54CAE0(entry) != 64 ||
		simplePolypOriginalSimpleRuntimeOverride54CAE0 != nil || simplePolypOriginalRuntimeOverride54CB10 != nil || simplePolypOriginalStateOverride54CB10 != nil {
		return 0, false
	}
	return simplePolypOriginalC54CAE0(entry, source), true
}

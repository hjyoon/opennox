package legacy

import "github.com/opennox/opennox/v1/server"

// SpecialDeathActualDefaultC54E460 uses the real public C symbol and all
// production world services. No original-reference runtime/state is installed.
func SpecialDeathActualDefaultC54E460(entry int, unit *server.Object) bool {
	if entry < 0 || entry >= 3 || specialDeathOriginalABI54E460(entry) != 64 || specialDeathOriginalRuntimeOverride54E460 != nil || specialDeathOriginalStateOverride54E460 != nil {
		return false
	}
	specialDeathOriginalC54E460(entry, unit)
	return true
}

func SpecialDeathActualPlayerWords54E460(unit *server.Object) [4]uint32 {
	_, words := specialDeathOriginalPlayer54E460(unit)
	return words
}

package legacy

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

// SkeletonDeathActualDefaultC54A310 exercises a public C entry without
// reference-service overrides. Entry numbers are Skeleton/drop/debris/lord.
func SkeletonDeathActualDefaultC54A310(entry int, unit *server.Object, name string, mods [4]string, charge int32) (int32, bool) {
	if entry < 0 || entry > 3 || skeletonOriginalABI54A310(entry) != 64 || skeletonOriginalRuntimeOverride54A310 != nil || skeletonOriginalStateOverride54A310 != nil {
		return 0, false
	}
	var names [4]unsafe.Pointer
	for i, s := range mods {
		if s != "" {
			names[i] = unsafe.Pointer(internCStr(s))
		}
	}
	return skeletonOriginalC54A310(entry, unit, unsafe.Pointer(internCStr(name)), names, charge), true
}

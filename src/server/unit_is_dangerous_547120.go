package server

import "github.com/opennox/libs/object"

type dangerousUnitTypeCache547120 struct {
	toxicCloud      uint32
	smallToxicCloud uint32
}

type dangerousUnitHooks547120[O any] struct {
	loadToxicCloudCache       func() uint32
	lookupType                func(string) uint32
	storeToxicCloudCache      func(uint32)
	storeSmallToxicCloudCache func(uint32)
	loadClass                 func(O) uint32
	loadType                  func(O) uint16
	loadSmallToxicCloudCache  func() uint32
	loadSubClass              func(O) uint32
	clearLocationSafe         func()
}

// dangerousUnit547120 reconstructs GAME.EXE 00547120. ToxicCloud alone is
// the lazy-initialization sentinel for both type caches. Candidate class is
// observed before type, the ToxicCloud cache is reloaded for the comparison,
// and SmallToxicCloud is loaded only after that comparison fails. Fire-class
// candidates use unit subclass bit 0x400; toxic clouds use bit 0x200; every
// other Dangerous-class candidate clears the caller-owned location-safe flag.
//
// The original callback's return is not a Boolean: it is 0/1 for Fire, the
// full unit subclass for either cloud, and the candidate type otherwise. Its
// only production caller discards this value, but retaining it here preserves
// the complete PE32 contract for oracle tests.
func dangerousUnit547120[O any](candidate, unit O, hooks dangerousUnitHooks547120[O]) uint32 {
	toxicCloud := hooks.loadToxicCloudCache()
	if toxicCloud == 0 {
		toxicCloud = hooks.lookupType("ToxicCloud")
		hooks.storeToxicCloudCache(toxicCloud)
		smallToxicCloud := hooks.lookupType("SmallToxicCloud")
		hooks.storeSmallToxicCloudCache(smallToxicCloud)
	}

	class := hooks.loadClass(candidate)
	if class&uint32(object.ClassFire) != 0 {
		result := (hooks.loadSubClass(unit) >> 10) & 1
		if result == 0 {
			hooks.clearLocationSafe()
		}
		return result
	}

	result := uint32(hooks.loadType(candidate))
	if result == hooks.loadToxicCloudCache() {
		result = hooks.loadSubClass(unit)
		if result&0x200 == 0 {
			hooks.clearLocationSafe()
		}
		return result
	}
	if result == hooks.loadSmallToxicCloudCache() {
		result = hooks.loadSubClass(unit)
		if result&0x200 == 0 {
			hooks.clearLocationSafe()
		}
		return result
	}
	if class&uint32(object.ClassDangerous) != 0 {
		hooks.clearLocationSafe()
	}
	return result
}

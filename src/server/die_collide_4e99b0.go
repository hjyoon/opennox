package server

const (
	dieCollideUnitClassMask4E99B0 = uint8(0x6)
	dieCollideDestroyedFlag4E99B0 = uint32(0x8000)
)

type dieCollideHooks4E99B0[O, D comparable] struct {
	unitsOnSameTeam func(O, O) int32
	classLow        func(O) uint8
	loadFlags       func(O) uint32
	loadDeath       func(O) D
	storeFlags      func(O, uint32)
	callDeath       func(D, O)
	delayedDelete   func(O)
}

// dieCollide4E99B0 preserves GAME.EXE 004E99B0. The target nil check occurs
// before the source argument is used. On the accepted path, source flags and
// the death callback are loaded in that order, the Destroyed bit is computed
// from the cached flags, and the flags are stored before either the cached
// death callback or delayed deletion runs. The registered collide-data size is
// zero and the third collision argument is not read.
func dieCollide4E99B0[O, D comparable](
	source, target O,
	collision any,
	hooks dieCollideHooks4E99B0[O, D],
) {
	_ = collision
	var zeroObject O
	if target == zeroObject {
		return
	}
	if hooks.unitsOnSameTeam(source, target) != 0 {
		return
	}
	if hooks.classLow(target)&dieCollideUnitClassMask4E99B0 == 0 {
		return
	}

	flags := hooks.loadFlags(source)
	death := hooks.loadDeath(source)
	hooks.storeFlags(source, flags|dieCollideDestroyedFlag4E99B0)

	var zeroDeath D
	if death != zeroDeath {
		hooks.callDeath(death, source)
		return
	}
	hooks.delayedDelete(source)
}

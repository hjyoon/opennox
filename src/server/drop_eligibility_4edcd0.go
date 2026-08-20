package server

const (
	dropEligibilityDestroyed4EDCD0 = uint32(0x00000020)
	dropEligibilityUnitClass4EDCD0 = uint8(0x06)
	dropEligibilityNoDrop4EDCD0    = uint32(0x10000000)
)

type dropEligibilityHooks4EDCD0[O, I any] struct {
	loadItemArg       func() I
	loadItemFlags     func(I) uint32
	loadOwnerArg      func() O
	loadOwnerClassLow func(O) uint8
}

// dropEligibility4EDCD0 preserves GAME.EXE 004EDCD0. The item and its full
// flags dword are read before the owner argument. A Destroyed item returns
// before touching the owner; otherwise only the owner's low class byte is
// read. Every true branch returns canonical one and the sole false case is a
// live item with NoDrop set while held by a Player or Monster owner.
func dropEligibility4EDCD0[O, I any](hooks dropEligibilityHooks4EDCD0[O, I]) int32 {
	item := hooks.loadItemArg()
	flags := hooks.loadItemFlags(item)
	if uint8(flags)&uint8(dropEligibilityDestroyed4EDCD0) != 0 {
		return 1
	}
	owner := hooks.loadOwnerArg()
	if hooks.loadOwnerClassLow(owner)&dropEligibilityUnitClass4EDCD0 == 0 {
		return 1
	}
	if flags&dropEligibilityNoDrop4EDCD0 == 0 {
		return 1
	}
	return 0
}

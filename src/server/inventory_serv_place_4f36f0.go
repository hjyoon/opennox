package server

const (
	inventoryServPlaceDestroyedFlagLow4F36F0 = uint8(0x20)
	inventoryServPlaceDeadFlag4F36F0         = uint32(0x00008000)
	inventoryServPlaceUnitClassLow4F36F0     = uint8(0x06)
	inventoryServPlaceNoCollideFlag4F36F0    = uint32(0x00000040)
)

// inventoryServPlaceHooks4F36F0 exposes every observable load, call, and
// store in GAME.EXE 004F36F0. Pickup and collision callbacks are separate
// comparable tokens so the original cached callback selection can be tested
// without turning a native pointer into a host-width integer.
type inventoryServPlaceHooks4F36F0[O, P, C comparable] struct {
	loadOwnerCarryCapacity func(O) uint16
	loadItemFlagsLow       func(O) uint8
	loadOwnerFlags         func(O) uint32
	loadItemType           func(O) uint16
	itemTypeAllowed        func(uint16) int32
	loadOwnerClassLow      func(O) uint8

	loadPickup    func(O) P
	loadArg4      func() int32
	loadArg3      func() int32
	callPickup    func(P, O, O, int32, int32) int32
	defaultPickup func(O, O, int32, int32) int32

	loadItemFlags  func(O) uint32
	storeItemFlags func(O, uint32)
	loadCollide    func(O) C
	refreshCollide func(O)

	loadScriptPickupFunc  func(O) int32
	callScriptPickup      func(O, O)
	storeScriptPickupFunc func(O, int32)
}

// inventoryServPlace4F36F0 preserves GAME.EXE 004F36F0.
//
// Owner and item admission follows the exact guard order: owner, item, carry
// capacity, item's low destroyed byte, owner's full dead flag, zero-extended
// item type admission, and owner's low unit-class byte. The item's pickup
// callback is cached before arg4 and arg3 are loaded. A nil callback uses
// DefaultPickup; either path returns its full int32 result, and zero skips all
// post-pickup work.
//
// A successful result reloads the item's full flags. NoCollide is cleared from
// that cached value before the live Collide pointer is read and an optional
// collision refresh is called. The pickup script Func is then read live; any
// value other than -1 invokes the script and is overwritten with -1 after the
// callback. Deliberately do not add nil guards beyond the two entry guards.
func inventoryServPlace4F36F0[O, P, C comparable](
	owner, item O,
	hooks inventoryServPlaceHooks4F36F0[O, P, C],
) int32 {
	var zeroObject O
	if owner == zeroObject {
		return 0
	}
	if item == zeroObject {
		return 0
	}
	if hooks.loadOwnerCarryCapacity(owner) == 0 {
		return 0
	}
	if hooks.loadItemFlagsLow(item)&inventoryServPlaceDestroyedFlagLow4F36F0 != 0 {
		return 0
	}
	if hooks.loadOwnerFlags(owner)&inventoryServPlaceDeadFlag4F36F0 != 0 {
		return 0
	}
	typeInd := hooks.loadItemType(item)
	if hooks.itemTypeAllowed(typeInd) == 0 {
		return 0
	}
	if hooks.loadOwnerClassLow(owner)&inventoryServPlaceUnitClassLow4F36F0 == 0 {
		return 0
	}

	pickup := hooks.loadPickup(item)
	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	var result int32
	var zeroPickup P
	if pickup != zeroPickup {
		result = hooks.callPickup(pickup, owner, item, arg3, arg4)
	} else {
		result = hooks.defaultPickup(owner, item, arg3, arg4)
	}
	if result == 0 {
		return result
	}

	flags := hooks.loadItemFlags(item)
	if flags&inventoryServPlaceNoCollideFlag4F36F0 != 0 {
		hooks.storeItemFlags(item, flags&^inventoryServPlaceNoCollideFlag4F36F0)
		collide := hooks.loadCollide(item)
		var zeroCollide C
		if collide != zeroCollide {
			hooks.refreshCollide(item)
		}
	}

	if hooks.loadScriptPickupFunc(item) != -1 {
		hooks.callScriptPickup(owner, item)
		hooks.storeScriptPickupFunc(item, -1)
	}
	return result
}

package server

const (
	playerRespawnItemClearFlag4EF750  uint32 = 0x00080000
	playerRespawnItemUpdateMask4EF750 uint32 = 0x03001000
)

type playerRespawnItemHooks4EF750[O, I, A, U comparable] struct {
	loadTypeIDArg   func() string
	newObject       func(string) O
	loadInit        func(O) I
	callInit        func(I, O, uint32)
	loadAttrsArg    func() A
	applyAttrs      func(O, A)
	loadPlaceA5Arg  func() int32
	loadPlaceA4Arg  func() int32
	loadPlayerArg   func() O
	placeInventory  func(O, O, int32, int32)
	loadFlags       func(O) uint32
	loadClass       func(O) uint32
	storeFlags      func(O, uint32)
	loadUpdateData  func(O) U
	loadUpdateMark  func(U) uint32
	storeUpdateMark func(U, uint32)
}

// playerRespawnItem4EF750 preserves GAME.EXE 004EF750. The type-ID argument
// is loaded before object creation. A nil creation result returns immediately
// without loading any other argument or object field.
//
// For a created object, the initializer pointer is loaded once and, when
// nonzero, called with the object and a fixed zero. The optional attributes
// argument is loaded only after that callback. Placement arguments are then
// loaded in the original a5, a4, player order and the placement result is
// ignored.
//
// Placement may mutate the item, so flags and class are loaded live afterward
// in that order. The cached flags have bit 0x00080000 cleared and are stored
// after the class test is computed. Only Wand, Weapon, or Armor class bits
// cause a subsequent live UpdateData load and an OR of bit zero into its
// dword at byte offset four. The created object is returned on every path.
func playerRespawnItem4EF750[O, I, A, U comparable](
	hooks playerRespawnItemHooks4EF750[O, I, A, U],
) O {
	typeID := hooks.loadTypeIDArg()
	item := hooks.newObject(typeID)
	var zeroObject O
	if item == zeroObject {
		return item
	}

	init := hooks.loadInit(item)
	var zeroInit I
	if init != zeroInit {
		hooks.callInit(init, item, 0)
	}

	attrs := hooks.loadAttrsArg()
	var zeroAttrs A
	if attrs != zeroAttrs {
		hooks.applyAttrs(item, attrs)
	}

	a5 := hooks.loadPlaceA5Arg()
	a4 := hooks.loadPlaceA4Arg()
	player := hooks.loadPlayerArg()
	hooks.placeInventory(player, item, a4, a5)

	flags := hooks.loadFlags(item)
	class := hooks.loadClass(item)
	flags &^= playerRespawnItemClearFlag4EF750
	markUpdate := class&playerRespawnItemUpdateMask4EF750 != 0
	hooks.storeFlags(item, flags)
	if markUpdate {
		update := hooks.loadUpdateData(item)
		mark := hooks.loadUpdateMark(update)
		hooks.storeUpdateMark(update, mark|1)
	}
	return item
}

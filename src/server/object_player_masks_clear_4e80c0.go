package server

type objectPlayerMasksClearHooks4E80C0[O comparable] struct {
	firstObject  func() O
	loadField36  func(O) uint32
	loadField35  func(O) uint32
	storeField36 func(O, uint32)
	storeField35 func(O, uint32)
	nextObject   func(O) O
}

// objectPlayerMasksClear4E80C0 preserves GAME.EXE 004E80C0. IA-32 masks
// the variable shift count to five bits. Each object caches Field36 and
// Field35, in that order, before storing both cleared masks in the same order.
func objectPlayerMasksClear4E80C0[O comparable](
	playerInd uint32,
	hooks objectPlayerMasksClearHooks4E80C0[O],
) O {
	bit := uint32(1) << (playerInd & 31)
	var zero O
	obj := hooks.firstObject()
	if obj == zero {
		return zero
	}
	clear := ^bit
	for obj != zero {
		field36 := hooks.loadField36(obj)
		field35 := hooks.loadField35(obj)
		hooks.storeField36(obj, field36&clear)
		hooks.storeField35(obj, field35&clear)
		obj = hooks.nextObject(obj)
	}
	return obj
}

package server

type objectPixieTargetClearResult4E81D0[U any] struct {
	typeID        uint32
	updateData    U
	returnsUpdate bool
}

type objectPixieTargetClearHooks4E81D0[O comparable, U any] struct {
	loadPixieTypeID  func() uint32
	lookupObjectType func(string) uint32
	storePixieTypeID func(uint32)
	loadTypeInd      func(O) uint16
	loadUpdateData   func(O) U
	clearTarget      func(U)
}

// objectPixieTargetClear4E81D0 preserves GAME.EXE 004E81D0. The Pixie type
// cache is loaded exactly once and, when empty, populated before the object is
// inspected. A matching object's update-data target is then cleared.
//
// The original EAX contains the type ID on a nil/mismatch path and the update
// data pointer on a match. The tagged result retains that distinction without
// converting a native pointer to an integer on 64-bit targets.
func objectPixieTargetClear4E81D0[O comparable, U any](
	obj O,
	hooks objectPixieTargetClearHooks4E81D0[O, U],
) objectPixieTargetClearResult4E81D0[U] {
	typeID := hooks.loadPixieTypeID()
	if typeID == 0 {
		typeID = hooks.lookupObjectType("Pixie")
		hooks.storePixieTypeID(typeID)
	}
	result := objectPixieTargetClearResult4E81D0[U]{typeID: typeID}
	var zero O
	if obj == zero || uint32(hooks.loadTypeInd(obj)) != typeID {
		return result
	}
	updateData := hooks.loadUpdateData(obj)
	hooks.clearTarget(updateData)
	result.updateData = updateData
	result.returnsUpdate = true
	return result
}

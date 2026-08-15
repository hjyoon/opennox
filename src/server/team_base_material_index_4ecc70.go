package server

const teamBaseTypeName4ECC70 = "TeamBase"

// teamBaseMaterialIndexHooks4ECC70 keeps the cached object-type ID, object,
// init-data, modifier, and material lookup domains explicit. GAME.EXE stores
// the type ID in a 32-bit BSS cell but reads Object.TypeInd as an unsigned
// 16-bit value.
type teamBaseMaterialIndexHooks4ECC70[O, D, M, V any] struct {
	loadCachedType     func() uint32
	lookupType         func(string) uint32
	storeCachedType    func(uint32)
	loadTypeIndex      func(O) uint16
	loadInitData       func(O) D
	loadSecondModifier func(D) M
	lookupMaterial     func(M) V
}

// teamBaseMaterialIndex4ECC70 preserves GAME.EXE 004ECC70. The cached
// TeamBase type ID is loaded before the object is touched. A zero cache calls
// the name lookup and stores its result, including zero, before Object.TypeInd
// is read. The returned lookup value remains live for the comparison; the BSS
// cell is not reloaded after the store.
//
// A type mismatch returns zero without loading init-data. A match live-loads
// init-data, then its second modifier pointer, and passes that value (including
// nil) to the private 004ECC00 material lookup.
//
// GAME.EXE and retained OpenNox history contain no call, jump, or stored
// function-pointer reference to this routine, so no public or CGo edge is
// invented here.
func teamBaseMaterialIndex4ECC70[O, D, M, V any](
	obj O,
	hooks teamBaseMaterialIndexHooks4ECC70[O, D, M, V],
) V {
	typeIndex := hooks.loadCachedType()
	if typeIndex == 0 {
		typeIndex = hooks.lookupType(teamBaseTypeName4ECC70)
		hooks.storeCachedType(typeIndex)
	}

	if uint32(hooks.loadTypeIndex(obj)) != typeIndex {
		var zero V
		return zero
	}
	data := hooks.loadInitData(obj)
	modifier := hooks.loadSecondModifier(data)
	return hooks.lookupMaterial(modifier)
}

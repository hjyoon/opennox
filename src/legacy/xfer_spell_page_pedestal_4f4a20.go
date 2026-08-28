package legacy

const spellPagePedestalXferCurrentVersion4F4A20 = uint16(60)

// spellPagePedestalXferDeps4F4A20 exposes every observable object access,
// transfer, and global read in GAME.EXE 004F4A20. Object and collide-data
// identities remain generic so the contract cannot inherit PE32 pointer width.
type spellPagePedestalXferDeps4F4A20[O comparable, D any] struct {
	loadField34       func(O) uint32
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	loadCollideData   func(O) D
	rwSpellPayload    func(D)
	readOnly          func() int32
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// spellPagePedestalXfer4F4A20 preserves the original entry-time Field34
// cache, signed version check, live CollideData load, unconditional four-byte
// spell transfer, live inventory-count load, exact-one read gate, and failure
// prefixes. There is deliberately no nil guard for either object or payload.
func spellPagePedestalXfer4F4A20[O comparable, D any](
	object O,
	deps spellPagePedestalXferDeps4F4A20[O, D],
) int32 {
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(spellPagePedestalXferCurrentVersion4F4A20)
	version := int16(versionWord)
	if version > int16(spellPagePedestalXferCurrentVersion4F4A20) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	spellPayload := deps.loadCollideData(object)
	deps.rwSpellPayload(spellPayload)

	inventoryCount := deps.loadField34(object)
	if inventoryCount != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

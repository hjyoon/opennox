package legacy

const readableXferCurrentVersion4F4AB0 = uint16(60)

// readableXferDeps4F4AB0 exposes every observable object/use-data access,
// transfer, and global read in GAME.EXE 004F4AB0. Object and use-data
// identities remain generic so the contract cannot inherit PE32 pointer width.
type readableXferDeps4F4AB0[O comparable, D any] struct {
	loadUseData        func(O) D
	textSizeWithNUL    func(D) uint32
	loadField34        func(O) uint32
	rwVersion          func(uint16) uint16
	mapReadWrite       func(O, int32) int32
	rwTextSize         func(uint32) uint32
	rwText             func(D, uint32)
	readOnly           func() int32
	clearTransientRead func(D)
	transferInventory  func(uint16, O, int32) int32
	storeField34       func(O, uint32)
}

// readableXfer4F4AB0 preserves the original entry-time UseData/text-size and
// Field34 caches, signed version branches, unbounded text transfer, exact-one
// read gate, live inventory-count load, and failure prefixes. There is
// deliberately no nil or text-size guard: both are observable in GAME.EXE.
func readableXfer4F4AB0[O comparable, D any](
	object O,
	deps readableXferDeps4F4AB0[O, D],
) int32 {
	useData := deps.loadUseData(object)
	textSize := deps.textSizeWithNUL(useData)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(readableXferCurrentVersion4F4AB0)
	version := int16(versionWord)
	if version > int16(readableXferCurrentVersion4F4AB0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if version >= 2 {
		textSize = deps.rwTextSize(textSize)
	}
	deps.rwText(useData, textSize)

	if deps.readOnly() == 1 {
		deps.clearTransientRead(useData)
		inventoryCount := deps.loadField34(object)
		if inventoryCount != 0 {
			if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
				return 0
			}
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

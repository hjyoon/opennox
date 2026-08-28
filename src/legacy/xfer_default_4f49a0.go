package legacy

const defaultXferCurrentVersion4F49A0 = uint16(60)

// defaultXferDeps4F49A0 exposes every observable object access, transfer, and
// global read in GAME.EXE 004F49A0. Object identity remains generic so the
// contract cannot inherit the original PE32 pointer width.
type defaultXferDeps4F49A0[O comparable] struct {
	loadField34       func(O) uint32
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	readOnly          func() int32
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// defaultXfer4F49A0 preserves the original entry-time Field34 cache, signed
// version check, live inventory-count load, exact-one read gate, and failure
// prefixes. There is deliberately no nil-object guard. The callback's second
// argument is absent because GAME.EXE never reads it.
func defaultXfer4F49A0[O comparable](
	object O,
	deps defaultXferDeps4F49A0[O],
) int32 {
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(defaultXferCurrentVersion4F49A0)
	version := int16(versionWord)
	if version > int16(defaultXferCurrentVersion4F49A0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	inventoryCount := deps.loadField34(object)
	if inventoryCount != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

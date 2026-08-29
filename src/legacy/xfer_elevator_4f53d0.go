package legacy

const (
	elevatorXferCurrentVersion4F53D0 = uint16(61)
	elevatorXferField4Version4F53D0  = int16(41)
	elevatorXferField3Version4F53D0  = int16(61)
)

// elevatorXferDeps4F53D0 exposes every observable object/update-data access,
// transfer, global-mode read, and external call in GAME.EXE 004F53D0.
// Object and update-data identities remain generic so this contract cannot
// inherit PE32 pointer width.
type elevatorXferDeps4F53D0[O comparable, D any] struct {
	loadUpdateData    func(O) D
	loadField34       func(O) uint32
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	rwShaftExtent     func(D)
	rwField4          func(D)
	rwField3          func(D)
	readOnly          func() int32
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// elevatorXfer4F53D0 preserves the entry-time UpdateData and Field34 caches,
// signed version checks and branches, live inventory-count load, exact-one
// inventory gate, zero-extended inventory version, and failure prefixes.
// There are deliberately no object or update-data guards.
func elevatorXfer4F53D0[O comparable, D any](
	object O,
	deps elevatorXferDeps4F53D0[O, D],
) int32 {
	data := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(elevatorXferCurrentVersion4F53D0)
	version := int16(versionWord)
	if version > int16(elevatorXferCurrentVersion4F53D0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwShaftExtent(data)
	if version >= elevatorXferField4Version4F53D0 {
		deps.rwField4(data)
	}
	if version >= elevatorXferField3Version4F53D0 {
		deps.rwField3(data)
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

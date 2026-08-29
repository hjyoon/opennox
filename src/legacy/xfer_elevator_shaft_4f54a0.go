package legacy

const elevatorShaftXferCurrentVersion4F54A0 = uint16(60)

// elevatorShaftXferDeps4F54A0 exposes every observable object/update-data
// access, transfer, global-mode read, and external call in GAME.EXE 004F54A0.
// Object and update-data identities remain generic so this contract cannot
// inherit PE32 pointer width.
type elevatorShaftXferDeps4F54A0[O comparable, D any] struct {
	loadUpdateData    func(O) D
	loadField34       func(O) uint32
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	rwElevatorExtent  func(D)
	readOnly          func() int32
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// elevatorShaftXfer4F54A0 preserves the entry-time UpdateData and Field34
// caches, signed version check, live inventory-count load, exact-one inventory
// gate, zero-extended inventory version, and original failure prefixes. There
// are deliberately no object or update-data guards.
func elevatorShaftXfer4F54A0[O comparable, D any](
	object O,
	deps elevatorShaftXferDeps4F54A0[O, D],
) int32 {
	data := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(elevatorShaftXferCurrentVersion4F54A0)
	version := int16(versionWord)
	if version > int16(elevatorShaftXferCurrentVersion4F54A0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwElevatorExtent(data)
	inventoryCount := deps.loadField34(object)
	if inventoryCount != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

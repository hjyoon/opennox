package legacy

const transporterXferCurrentVersion4F5300 = uint16(60)

// transporterXferDeps4F5300 exposes every observable object/update-data
// access, transfer, global-mode read, and external call in GAME.EXE 004F5300.
// Object and update-data identities remain generic so this contract cannot
// inherit PE32 pointer width.
type transporterXferDeps4F5300[O comparable, D any] struct {
	loadUpdateData      func(O) D
	loadField34         func(O) uint32
	rwVersion           func(uint16) uint16
	mapReadWrite        func(O, int32) int32
	readOnly            func() int32
	rwTargetExtent      func(D)
	hasTarget           func(D) bool
	loadTargetExtent    func(D) uint32
	rwLocalTargetExtent func(uint32)
	transferInventory   func(uint16, O, int32) int32
	storeField34        func(O, uint32)
}

// transporterXfer4F5300 preserves the entry-time UpdateData and Field34
// caches, signed version check, nonzero read-mode extent transfer, write-mode
// target gate and local extent copy, live inventory-count load, exact-one
// inventory gate, zero-extended inventory version, and failure prefixes.
// There are deliberately no object or update-data guards.
func transporterXfer4F5300[O comparable, D any](
	object O,
	deps transporterXferDeps4F5300[O, D],
) int32 {
	data := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(transporterXferCurrentVersion4F5300)
	version := int16(versionWord)
	if version > int16(transporterXferCurrentVersion4F5300) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readOnly() != 0 {
		deps.rwTargetExtent(data)
	} else {
		var targetExtent uint32
		if deps.hasTarget(data) {
			targetExtent = deps.loadTargetExtent(data)
		}
		deps.rwLocalTargetExtent(targetExtent)
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

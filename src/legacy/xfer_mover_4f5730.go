package legacy

const moverXferCurrentVersion4F5730 = uint16(60)

// moverXferDeps4F5730 exposes every observable object/update-data access,
// transfer, mode read, waypoint lookup, and external call in GAME.EXE
// 004F5730. Object and update-data identities remain generic so this contract
// cannot inherit PE32 pointer width.
type moverXferDeps4F5730[O comparable, D any] struct {
	loadUpdateData func(O) D
	loadField34    func(O) uint32
	rwVersion      func(uint16) uint16
	mapReadWrite   func(O, int32) int32

	rwField1 func(D)
	rwField2 func(D)
	rwField8 func(D)
	rwField0 func(D)

	readOnly        func() int32
	rwField4        func(D)
	rwField6        func(D)
	waypointIndex   func(O, D, int) uint32
	rwWaypointIndex func(uint32)

	rwSpeedBase func(O)
	rwSpeedCur  func(O)

	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// moverXfer4F5730 preserves the entry-time UpdateData and Field34 caches,
// signed version thresholds, write-side waypoint dereferences, live
// inventory-count load, exact-one inventory gate, zero-extended inventory
// version, and original failure prefixes. There are deliberately no object or
// update-data guards.
func moverXfer4F5730[O comparable, D any](
	object O,
	deps moverXferDeps4F5730[O, D],
) int32 {
	data := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(moverXferCurrentVersion4F5730)
	version := int16(versionWord)
	if version > int16(moverXferCurrentVersion4F5730) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwField1(data)
	deps.rwField2(data)
	deps.rwField8(data)
	if version >= 41 {
		deps.rwField0(data)
		if deps.readOnly() != 0 {
			deps.rwField4(data)
			deps.rwField6(data)
		} else {
			deps.rwWaypointIndex(deps.waypointIndex(object, data, 3))
			deps.rwWaypointIndex(deps.waypointIndex(object, data, 5))
		}
	}
	if version >= 42 {
		deps.rwSpeedBase(object)
		deps.rwSpeedCur(object)
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

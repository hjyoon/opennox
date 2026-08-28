package legacy

const exitXferCurrentVersion4F4B90 = uint16(60)

// exitXferDeps4F4B90 exposes every observable object/collide-data access,
// transfer, and read-only query in GAME.EXE 004F4B90. Object and data
// identities remain generic so the contract cannot inherit PE32 pointer width.
type exitXferDeps4F4B90[O comparable, D any] struct {
	loadCollideData     func(O) D
	mapNameSizeWithNUL  func(D) uint32
	loadField34         func(O) uint32
	rwVersion           func(uint16) uint16
	mapReadWrite        func(O, int32) int32
	readOnly            func() int32
	rwMapNameSize       func(uint32) uint32
	rwMapName           func(D, uint32)
	rwLegacyMapNameByte func(D, uint32)
	loadMapNameByte     func(D, uint32) byte
	rwDestinationX      func(D)
	rwDestinationY      func(D)
	transferInventory   func(uint16, O, int32) int32
	storeField34        func(O, uint32)
}

// exitXfer4F4B90 preserves the original entry-time CollideData/map-name-size
// and Field34 caches, signed version branches, unbounded old/new map-name
// transfers, cached destination payload, exact-one inventory gate, live count,
// and failure prefixes. There is deliberately no object, data, or size guard.
func exitXfer4F4B90[O comparable, D any](
	object O,
	deps exitXferDeps4F4B90[O, D],
) int32 {
	data := deps.loadCollideData(object)
	mapNameSize := deps.mapNameSizeWithNUL(data)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(exitXferCurrentVersion4F4B90)
	version := int16(versionWord)
	if version > int16(exitXferCurrentVersion4F4B90) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if version < 2 {
		if deps.readOnly() == 1 {
			for offset := uint32(0); ; offset++ {
				deps.rwLegacyMapNameByte(data, offset)
				if deps.loadMapNameByte(data, offset) == 0 {
					break
				}
			}
		} else {
			deps.rwMapName(data, mapNameSize)
		}
	} else {
		mapNameSize = deps.rwMapNameSize(mapNameSize)
		deps.rwMapName(data, mapNameSize)
	}

	if version >= 31 {
		deps.rwDestinationX(data)
		deps.rwDestinationY(data)
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

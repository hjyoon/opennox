package legacy

const (
	holeXferCurrentVersion4F51D0 = uint16(60)
	holeXferField24Version4F51D0 = int16(42)
	holeXferScriptVersion4F51D0  = int16(41)
)

// holeXferDeps4F51D0 exposes every observable object/collide-data access,
// transfer, global-mode read, and external call in GAME.EXE 004F51D0.
// Object, collide-data, and script-data identities remain generic so this
// contract cannot inherit PE32 pointer width.
type holeXferDeps4F51D0[O comparable, D any, S any] struct {
	loadField34             func(O) uint32
	loadScriptData          func(O) S
	loadCollideData         func(O) D
	rwVersion               func(uint16) uint16
	mapReadWrite            func(O, int32) int32
	rwField24               func(D)
	storeField24            func(D, uint32)
	transferScript          func(D, S, uintptr)
	rwDestinationXY         func(D)
	rwDestinationExtent     func(D)
	rwDestinationNetCode    func(D)
	storeScriptFunc         func(D, int32)
	storeScriptFlags        func(D, uint32)
	storeDestinationExtent  func(D, uint32)
	storeDestinationNetCode func(D, uint16)
	readOnly                func() int32
	transferInventory       func(uint16, O, int32) int32
	storeField34            func(O, uint32)
}

// holeXfer4F51D0 preserves the original entry-time Field34, ScriptData, and
// CollideData caches; signed version branches; legacy initialization order;
// exact-one inventory gate; live Field34 read; zero-extended inventory
// version; and failure prefixes. There are deliberately no object or
// collide-data guards. The native adapter separately maps a null ScriptData
// cache to the null script callback context used by the original.
func holeXfer4F51D0[O comparable, D any, S any](
	object O,
	deps holeXferDeps4F51D0[O, D, S],
) int32 {
	originalField34 := deps.loadField34(object)
	scriptData := deps.loadScriptData(object)
	data := deps.loadCollideData(object)

	versionWord := deps.rwVersion(holeXferCurrentVersion4F51D0)
	version := int16(versionWord)
	if version > int16(holeXferCurrentVersion4F51D0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if version < holeXferField24Version4F51D0 {
		deps.storeField24(data, 0)
	} else {
		deps.rwField24(data)
	}

	if version < holeXferScriptVersion4F51D0 {
		deps.rwDestinationXY(data)
		deps.storeScriptFunc(data, -1)
		deps.storeScriptFlags(data, 0)
		deps.storeDestinationExtent(data, 0)
		deps.storeDestinationNetCode(data, 0)
	} else {
		deps.transferScript(data, scriptData, 128)
		deps.rwDestinationXY(data)
		deps.rwDestinationExtent(data)
		deps.rwDestinationNetCode(data)
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

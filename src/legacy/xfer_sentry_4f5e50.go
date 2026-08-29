package legacy

const (
	sentryXferCurrentVersion4F5E50 = uint16(61)
	sentryXferUpdateSize4F5E50     = 12
	sentryXferGameMask4F5E50       = uint32(0x00200000)
)

// sentryXferDeps4F5E50 exposes every observable object, update-data, stream,
// mode, game-flag, and inventory access in GAME.EXE 004F5E50. Object and
// update-data identities remain generic so the contract cannot inherit PE32
// pointer truncation.
type sentryXferDeps4F5E50[O, D any] struct {
	loadUpdateData func(O) D
	loadField34    func(O) uint32
	rwVersion      func(uint16) uint16
	mapReadWrite   func(O, int32) int32

	rwUpdateData   func(D, int)
	readMode       func() int32
	gameFlags      func(uint32) int32
	loadUpdateU32  func(D, int) uint32
	storeUpdateU32 func(D, int, uint32)

	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// sentryXfer4F5E50 preserves the entry-time UpdateData and Field34 caches,
// signed version gates, exact three-dword PE32 update record, two independent
// mode reads, exact-one game-flag gate, and the inventory-failure prefix.
// UpdateData deliberately has no nil guard: the original first faults at the
// offset-four stream transfer, after common object serialization succeeds.
func sentryXfer4F5E50[O, D any](
	object O,
	deps sentryXferDeps4F5E50[O, D],
) int32 {
	data := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(sentryXferCurrentVersion4F5E50)
	version := int16(versionWord)
	if version > int16(sentryXferCurrentVersion4F5E50) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwUpdateData(data, 4)
	deps.rwUpdateData(data, 8)
	mode := deps.readMode()
	if mode == 1 || deps.gameFlags(sentryXferGameMask4F5E50) == 1 {
		value := deps.loadUpdateU32(data, 4)
		deps.storeUpdateU32(data, 0, value)
	}
	if version >= int16(sentryXferCurrentVersion4F5E50) {
		deps.rwUpdateData(data, 0)
	}

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

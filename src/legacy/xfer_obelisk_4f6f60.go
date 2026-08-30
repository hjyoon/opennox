package legacy

const (
	obeliskXferCurrentVersion4F6F60 = uint16(61)
	obeliskXferQuestFlag4F6F60      = uint32(2048)
)

// obeliskXferDeps4F6F60 exposes every observable object, update-data,
// stream, mode, drawable-list, and inventory access in GAME.EXE 004F6F60.
// Object and drawable identities stay generic so this contract cannot inherit
// the original PE32 pointer width.
type obeliskXferDeps4F6F60[O, U any, D comparable] struct {
	loadUpdateData    func(O) U
	loadField34       func(O) uint32
	storeField34      func(O, uint32)
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	rwMana            func(U)
	readMode          func() int32
	loadMana          func(U) int32
	syncManaLevel     func(O, float32)
	gameFlags         func(uint32) int32
	loadExtent        func(O) uint32
	staticDrawable    func(uint32) D
	firstMinimap      func() D
	nextMinimap       func(D) D
	rwMinimapPresent  func(uint8) uint8
	transferInventory func(uint16, O, int32) int32
}

// obeliskXferManaLevel4F6F60 reproduces the original i386 arithmetic used by
// the read-side notification: the 80*mana product wraps to 32 bits, that
// wrapped value is divided as a signed int32 with truncation toward zero, and
// the quotient is rounded once to binary32.
func obeliskXferManaLevel4F6F60(mana int32) float32 {
	wrapped := int32(uint32(mana) * uint32(80))
	return float32(wrapped / 50)
}

// obeliskXfer4F6F60 preserves the exact PE32 transfer order while all
// pointer-bearing identities remain native-width. In particular it keeps:
//
//   - the entry UpdateData/Field34 cache order;
//   - a signed int16 version gate and sign-extended common-object version;
//   - the version-61 four-byte mana and quest-minimap suffix;
//   - independent exact-one mode reads for mana notification and inventory;
//   - live extent and minimap-list identity checks regardless of stream mode;
//   - the zero-extended version word passed to inventory transfer; and
//   - success-only restoration of the entry Field34 value.
//
// There are intentionally no object or UpdateData nil guards. Stream callback
// results are ignored, the minimap byte read result is unused, and an
// inventory failure returns after every preceding side effect without
// restoring Field34, exactly as in the original routine.
func obeliskXfer4F6F60[O, U any, D comparable](
	object O,
	deps obeliskXferDeps4F6F60[O, U, D],
) int32 {
	updateData := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(obeliskXferCurrentVersion4F6F60)
	version := int16(versionWord)
	if version > int16(obeliskXferCurrentVersion4F6F60) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if version >= int16(obeliskXferCurrentVersion4F6F60) {
		minimapPresent := uint8(0)
		deps.rwMana(updateData)
		if deps.readMode() == 1 {
			deps.syncManaLevel(object, obeliskXferManaLevel4F6F60(deps.loadMana(updateData)))
		}
		if deps.gameFlags(obeliskXferQuestFlag4F6F60) != 0 {
			var zero D
			drawable := deps.staticDrawable(deps.loadExtent(object))
			if drawable != zero {
				for current := deps.firstMinimap(); current != zero; current = deps.nextMinimap(current) {
					if current == drawable {
						minimapPresent = 1
						break
					}
				}
			}
		}
		_ = deps.rwMinimapPresent(minimapPresent)
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

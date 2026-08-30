package legacy

const toxicCloudXferCurrentVersion4F70A0 = uint16(61)

// toxicCloudXferDeps4F70A0 exposes every observable object, update-data,
// stream, mode, and inventory access in GAME.EXE 004F70A0. Object and
// update-data identities stay generic so this contract cannot inherit the
// original PE32 pointer width.
type toxicCloudXferDeps4F70A0[O, U any] struct {
	loadUpdateData    func(O) U
	loadField34       func(O) uint32
	storeField34      func(O, uint32)
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	rwDuration        func(U)
	readMode          func() int32
	transferInventory func(uint16, O, int32) int32
}

// toxicCloudXfer4F70A0 preserves the exact PE32 transfer order while all
// pointer-bearing identities remain native-width. In particular it keeps:
//
//   - the entry UpdateData/Field34 cache order;
//   - the signed int16 version range 1..61 and sign-extended common version;
//   - the cached four-byte ToxicCloudUpdateData duration transfer;
//   - a live Field34 load before the exact-one read-mode test;
//   - the zero-extended version word passed to inventory transfer; and
//   - success-only restoration of the entry Field34 value.
//
// There are intentionally no nil guards. Stream callback results are ignored,
// and an inventory failure returns after every preceding side effect without
// restoring Field34, exactly as in the original routine.
func toxicCloudXfer4F70A0[O, U any](
	object O,
	deps toxicCloudXferDeps4F70A0[O, U],
) int32 {
	updateData := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(toxicCloudXferCurrentVersion4F70A0)
	version := int16(versionWord)
	if version > int16(toxicCloudXferCurrentVersion4F70A0) || version <= 0 {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwDuration(updateData)

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

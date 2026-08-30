package legacy

const goldXferCurrentVersion4F6EC0 = uint16(60)

// goldXferDeps4F6EC0 exposes every observable object, stream, mode, and
// inventory access in GAME.EXE 004F6EC0. Object and InitData identities stay
// generic so this contract cannot inherit the original PE32 pointer width.
type goldXferDeps4F6EC0[O, D any] struct {
	loadInitData      func(O) D
	loadField34       func(O) uint32
	storeField34      func(O, uint32)
	rwVersion         func(uint16) uint16
	mapReadWrite      func(O, int32) int32
	rwGoldAmount      func(D)
	readMode          func() int32
	transferInventory func(uint16, O, int32) int32
}

// goldXfer4F6EC0 preserves the exact PE32 transfer order while all
// pointer-bearing identities remain native-width. In particular it keeps:
//
//   - the entry InitData/Field34 cache order;
//   - a signed int16 version gate and sign-extended common-object version;
//   - the cached four-byte GoldInitData transfer after the common prefix;
//   - a live Field34 load before the exact-one read-mode test;
//   - the zero-extended version word passed to inventory transfer; and
//   - success-only restoration of the entry Field34 value.
//
// There are intentionally no nil guards. Stream callback results are ignored,
// and an inventory failure returns after every preceding side effect without
// restoring Field34, exactly as in the original routine.
func goldXfer4F6EC0[O, D any](object O, deps goldXferDeps4F6EC0[O, D]) int32 {
	initData := deps.loadInitData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(goldXferCurrentVersion4F6EC0)
	version := int16(versionWord)
	if version > int16(goldXferCurrentVersion4F6EC0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwGoldAmount(initData)

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

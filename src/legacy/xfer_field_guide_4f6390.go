package legacy

const fieldGuideXferCurrentVersion4F6390 = uint16(60)

// fieldGuideXferDeps4F6390 exposes every observable object, use-data,
// stream, mode, and inventory access in GAME.EXE 004F6390. Object and
// use-data identities remain generic so the contract cannot inherit PE32
// pointer truncation.
type fieldGuideXferDeps4F6390[O, D any] struct {
	loadUseData  func(O) D
	loadField34  func(O) uint32
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	readMode                func() int32
	creatureLength          func(D) uint32
	rwByte                  func(uint8) uint8
	rwCreature              func(D, uint8)
	storeCreatureTerminator func(D, uint8)

	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// fieldGuideXfer4F6390 preserves the PE32 entry cache order, signed version
// gate, two independently observed mode values, low-byte write length,
// unsigned 64-byte read limit, trailing-NUL store, live Field34 reload,
// exact-one inventory gate, zero-extended inventory version, and failure
// prefixes.
//
// Object and UseData deliberately have no nil guards. On write, the original
// performs an unbounded strlen from cached UseData and transfers the low byte
// of that length. On read, it transfers a one-byte length before touching
// cached UseData and stores a terminator even for an empty payload.
func fieldGuideXfer4F6390[O, D any](
	object O,
	deps fieldGuideXferDeps4F6390[O, D],
) int32 {
	data := deps.loadUseData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(fieldGuideXferCurrentVersion4F6390)
	version := int16(versionWord)
	if version > int16(fieldGuideXferCurrentVersion4F6390) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readMode() == 0 {
		size := deps.rwByte(uint8(deps.creatureLength(data)))
		deps.rwCreature(data, size)
	} else {
		size := deps.rwByte(0)
		if size >= 64 {
			return 0
		}
		deps.rwCreature(data, size)
		deps.storeCreatureTerminator(data, size)
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

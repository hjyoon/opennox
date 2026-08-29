package legacy

const (
	spellRewardXferCurrentVersion4F5F30 = uint16(60)
	spellRewardXferNameLimit4F5F30      = 128
	spellRewardXferOldIDLimit4F5F30     = uint8(0x89)
)

// spellRewardXferDeps4F5F30 exposes every observable object, use-data,
// stream, mode, spell-table, and inventory access in GAME.EXE 004F5F30.
// Object and use-data identities remain generic so the contract cannot
// inherit PE32 pointer truncation.
type spellRewardXferDeps4F5F30[O, D any] struct {
	loadField34  func(O) uint32
	loadUseData  func(O) D
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	readMode   func() int32
	rwByte     func(uint8) uint8
	rwBytes    func([]byte)
	loadSpell  func(D) uint8
	storeSpell func(D, uint8)
	spellName  func(uint8) string
	spellID    func(string) uint8

	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// spellRewardXferReadName4F5F30 reproduces the exact one-byte length gate.
// The original still invokes the byte transfer and name lookup for an empty
// name, so neither operation is skipped here.
func spellRewardXferReadName4F5F30[O, D any](
	deps spellRewardXferDeps4F5F30[O, D],
) (uint8, bool) {
	size := deps.rwByte(0)
	if size >= spellRewardXferNameLimit4F5F30 {
		return 0, false
	}
	buf := make([]byte, int(size))
	deps.rwBytes(buf)
	return deps.spellID(string(buf)), true
}

// spellRewardXferWriteName4F5F30 models the original 128-byte stack buffer.
// Registered spell names always fit. A longer runtime name crossed an
// undefined strcpy boundary in PE32, so the native contract faults instead
// of inventing a successful wire encoding or a rollback result.
func spellRewardXferWriteName4F5F30[O, D any](
	data D,
	deps spellRewardXferDeps4F5F30[O, D],
) {
	name := deps.spellName(deps.loadSpell(data))
	if len(name) >= spellRewardXferNameLimit4F5F30 {
		panic("SpellRewardXfer spell name exceeds the PE32 stack buffer")
	}
	var buf [spellRewardXferNameLimit4F5F30]byte
	copy(buf[:], name)
	size := deps.rwByte(uint8(len(name)))
	deps.rwBytes(buf[:int(size)])
}

// spellRewardXfer4F5F30 preserves the PE32 cache order, signed version gate,
// exact-one read mode, historical spell encodings, delayed legacy-name store,
// second live mode read, zero-extended inventory version, and failure prefixes.
// UseData deliberately has no nil guard: the original faults only when the
// selected payload path first dereferences the cached pointer, after common
// object serialization and its first mode read have succeeded.
func spellRewardXfer4F5F30[O, D any](
	object O,
	deps spellRewardXferDeps4F5F30[O, D],
) int32 {
	originalField34 := deps.loadField34(object)
	data := deps.loadUseData(object)

	versionWord := deps.rwVersion(spellRewardXferCurrentVersion4F5F30)
	version := int16(versionWord)
	if version > int16(spellRewardXferCurrentVersion4F5F30) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readMode() != 1 {
		spellRewardXferWriteName4F5F30(data, deps)
	} else if version < 31 {
		deps.rwByte(0) // First historical spell byte was intentionally ignored.
		second := deps.rwByte(0)
		third := deps.rwByte(0)
		if second >= spellRewardXferOldIDLimit4F5F30 {
			second = 0
		}
		if third >= spellRewardXferOldIDLimit4F5F30 {
			third = 0
		}
		deps.storeSpell(data, second)
		if third != 0 {
			deps.storeSpell(data, third)
		}
		if versionWord == 10 {
			deps.rwByte(0)
		}
	} else if version < 41 {
		if _, ok := spellRewardXferReadName4F5F30(deps); !ok {
			return 0
		}
		second, ok := spellRewardXferReadName4F5F30(deps)
		if !ok {
			return 0
		}
		third, ok := spellRewardXferReadName4F5F30(deps)
		if !ok {
			return 0
		}
		deps.storeSpell(data, second)
		if third != 0 {
			deps.storeSpell(data, third)
		}
	} else {
		value, ok := spellRewardXferReadName4F5F30(deps)
		if !ok {
			return 0
		}
		deps.storeSpell(data, value)
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

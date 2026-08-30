package legacy

const (
	monsterGeneratorXferCurrentVersion4F7130 = uint16(63)
	monsterGeneratorXferGroups4F7130         = 3
	monsterGeneratorXferSlots4F7130          = 4
)

type monsterGeneratorScriptSlot4F7130 uint8

const (
	monsterGeneratorScript48_4F7130 monsterGeneratorScriptSlot4F7130 = iota
	monsterGeneratorScript56_4F7130
	monsterGeneratorScript72_4F7130
	monsterGeneratorScript64_4F7130
)

type monsterGeneratorXferDeps4F7130[O comparable, U, S any] struct {
	loadUpdateData func(O) U
	loadField34    func(O) uint32
	storeField34   func(O, uint32)
	loadScriptData func(O) S

	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	rwSpawnSelectorCount func(uint8) uint8
	rwSpawnSelector      func(U, int)
	rwActiveCount        func(U)
	rwMaxActive          func(U)
	rwFrame88            func(U)
	transferScript       func(U, monsterGeneratorScriptSlot4F7130, S, uintptr) int32

	readMode              func() int32
	rwPrototypeGroupCount func(uint8) uint8
	loadPrototype         func(U, int) O
	rwPrototypeCount      func(uint8) uint8
	loadTypeName          func(O) []byte
	rwNameLength          func(uint8) uint8
	rwNameBytes           func([]byte)
	saveObject            func(O) int32
	rwPrototypeTag        func(uint16) uint16
	readPrototypeCRC      func()
	newObjectByTypeName   func([]byte) O
	callObjectXfer        func(O) int32
	storePrototype        func(U, int, O)

	rwQuestSelectorCount func(uint8) uint8
	rwQuestSelector      func(U, int)
	rwField92            func(U)
	transferInventory    func(uint16, O, int32) int32
}

// monsterGeneratorXfer4F7130 preserves GAME.EXE 004F7130's transfer order
// while keeping every object, update-data, and script-data identity at the
// native pointer width. The serialized format remains the original PE32
// format: object references are recursively transferred records, never raw
// native pointers.
//
// The selector and prototype counts are deliberately not clamped. GAME.EXE
// trusts them and advances beyond its nominal 3-by-4 arrays on malformed
// input. Native adapters retain that fault boundary through ordinary Go array
// indexing instead of silently accepting a different wire format.
//
// Stream and script-handler return values ignored by GAME.EXE are also ignored
// here. Only the common object prefix, object allocation/xfer, and suffix
// inventory operations can return failure. Field34 is restored to its entry
// value on success only.
func monsterGeneratorXfer4F7130[O comparable, U, S any](
	object O,
	deps monsterGeneratorXferDeps4F7130[O, U, S],
) int32 {
	updateData := deps.loadUpdateData(object)
	originalField34 := deps.loadField34(object)
	scriptData := deps.loadScriptData(object)

	versionWord := deps.rwVersion(monsterGeneratorXferCurrentVersion4F7130)
	version := int16(versionWord)
	if version <= 0 || version > int16(monsterGeneratorXferCurrentVersion4F7130) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	spawnSelectors := deps.rwSpawnSelectorCount(monsterGeneratorXferGroups4F7130)
	for i := 0; i < int(spawnSelectors); i++ {
		deps.rwSpawnSelector(updateData, i)
	}
	deps.rwActiveCount(updateData)
	deps.rwMaxActive(updateData)
	deps.rwFrame88(updateData)

	deps.transferScript(updateData, monsterGeneratorScript48_4F7130, scriptData, 1920)
	deps.transferScript(updateData, monsterGeneratorScript56_4F7130, scriptData, 2048)
	deps.transferScript(updateData, monsterGeneratorScript72_4F7130, scriptData, 2176)
	deps.transferScript(updateData, monsterGeneratorScript64_4F7130, scriptData, 2304)

	var zero O
	if deps.readMode() != 0 {
		groups := deps.rwPrototypeGroupCount(0)
		for group := 0; group < int(groups); group++ {
			count := deps.rwPrototypeCount(0)
			for slot := 0; slot < int(count); slot++ {
				length := deps.rwNameLength(0)
				name := make([]byte, int(length)+1)
				deps.rwNameBytes(name[:length])
				created := deps.newObjectByTypeName(name)
				if created == zero {
					return 0
				}
				deps.rwPrototypeTag(0)
				deps.readPrototypeCRC()
				if deps.callObjectXfer(created) == 0 {
					return 0
				}
				deps.storePrototype(updateData, group*monsterGeneratorXferSlots4F7130+slot, created)
			}
		}
	} else {
		// The transferred group count is metadata only on writes. GAME.EXE
		// always emits the fixed 3-by-4 in-memory prototype table.
		deps.rwPrototypeGroupCount(monsterGeneratorXferGroups4F7130)
		for group := 0; group < monsterGeneratorXferGroups4F7130; group++ {
			base := group * monsterGeneratorXferSlots4F7130
			var count uint8
			for slot := 0; slot < monsterGeneratorXferSlots4F7130; slot++ {
				if deps.loadPrototype(updateData, base+slot) != zero {
					count++
				}
			}
			deps.rwPrototypeCount(count)
			for slot := 0; slot < monsterGeneratorXferSlots4F7130; slot++ {
				prototype := deps.loadPrototype(updateData, base+slot)
				if prototype == zero {
					continue
				}
				firstName := deps.loadTypeName(prototype)
				length := deps.rwNameLength(uint8(len(firstName)))
				secondName := deps.loadTypeName(prototype)
				deps.rwNameBytes(secondName[:length])
				deps.saveObject(prototype)
			}
		}
	}

	if version >= 62 {
		questSelectors := deps.rwQuestSelectorCount(monsterGeneratorXferGroups4F7130)
		for i := 0; i < int(questSelectors); i++ {
			deps.rwQuestSelector(updateData, i)
		}
	}
	if version >= 63 {
		deps.rwField92(updateData)
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

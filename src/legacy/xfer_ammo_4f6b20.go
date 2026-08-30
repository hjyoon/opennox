package legacy

const ammoXferCurrentVersion4F6B20 = uint16(60)

// ammoXferDeps4F6B20 exposes every observable object, use-data, modifier,
// stream, mode, quest-flag, and inventory access in GAME.EXE 004F6B20.
// Pointer-bearing identities remain generic so this contract cannot inherit
// the original PE32 pointer width.
type ammoXferDeps4F6B20[O, I any, S comparable, M, U any] struct {
	loadUseData  func(O) U
	loadField34  func(O) uint32
	storeField34 func(O, uint32)
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	readMode           func() int32
	loadModifierData   func(O) I
	loadModifierSlot   func(I, int) S
	modifierNameLength func(S) uint32
	rwByte             func(uint8) uint8
	rwModifierName     func(S, uint8)
	readModifierName   func(uint8) string
	modifierIDByName   func(string) int32
	modifierByID       func(int32) M
	applyModifiers     func(O, [4]M, uint32)

	loadUseByte  func(U, int) uint8
	storeUseByte func(U, int, uint8)
	gameFlag4096 func() bool

	transferInventory func(uint16, O, int32) int32
}

// ammoXfer4F6B20 preserves the original PE32 transfer semantics while all
// object-owned identities remain native-width. In particular it preserves:
//
//   - entry UseData/Field34 cache order and the signed version gate;
//   - one modifier-direction read, four low-byte modifier-name transfers,
//     and the live write-slot reload after each length callback;
//   - cached AmmoUseData wire order byte 1 then byte 0;
//   - the read-only quest reset of byte 2 before charge stores;
//   - a live Field34 reload, exact-one inventory gate, and success-only
//     Field34 restoration.
//
// There are intentionally no nil guards. The original faults at the first
// dereference reached in an invalid object-owned record. Stream callback
// return values are ignored just as they were by the original routine.
func ammoXfer4F6B20[O, I any, S comparable, M, U any](
	object O,
	deps ammoXferDeps4F6B20[O, I, S, M, U],
) int32 {
	useData := deps.loadUseData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(ammoXferCurrentVersion4F6B20)
	version := int16(versionWord)
	if version > int16(ammoXferCurrentVersion4F6B20) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readMode() != 0 {
		var modifiers [4]M
		for i := range modifiers {
			size := deps.rwByte(0)
			// GAME.EXE compares this zero-extended byte with 256. The failure
			// edge is unreachable, so every possible byte length is accepted.
			name := deps.readModifierName(size)
			id := deps.modifierIDByName(name)
			modifiers[i] = deps.modifierByID(id)
		}
		deps.applyModifiers(object, modifiers, ^uint32(0))

		charge1 := deps.rwByte(0)
		charge0 := deps.rwByte(0)
		if deps.gameFlag4096() {
			deps.storeUseByte(useData, 2, 0)
		}
		deps.storeUseByte(useData, 1, charge1)
		deps.storeUseByte(useData, 0, charge0)
	} else {
		modifierData := deps.loadModifierData(object)
		var zero S
		for i := 0; i < 4; i++ {
			slot := deps.loadModifierSlot(modifierData, i)
			if slot == zero {
				deps.rwByte(0)
				continue
			}
			size := deps.rwByte(uint8(deps.modifierNameLength(slot)))
			// The descriptor/name is loaded again after the length callback.
			deps.rwModifierName(deps.loadModifierSlot(modifierData, i), size)
		}

		deps.rwByte(deps.loadUseByte(useData, 1))
		deps.rwByte(deps.loadUseByte(useData, 0))
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

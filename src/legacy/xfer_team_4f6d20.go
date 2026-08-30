package legacy

const teamXferCurrentVersion4F6D20 = uint16(60)

// teamXferDeps4F6D20 exposes every observable object, modifier, flag-update,
// stream, mode, and inventory access in GAME.EXE 004F6D20. Pointer-bearing
// identities remain generic so this contract cannot inherit PE32 width.
type teamXferDeps4F6D20[O, I any, S comparable, M, U any] struct {
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

	loadObjClass         func(O) uint32
	loadUpdateData       func(O) U
	loadPositionX        func(O) float32
	storeUpdatePositionX func(U, float32)
	loadPositionY        func(O) float32
	storeUpdatePositionY func(U, float32)

	transferInventory func(uint16, O, int32) int32
}

// teamXfer4F6D20 preserves the original PE32 transfer semantics while every
// object-owned identity remains native-width. In particular it preserves:
//
//   - the entry Field34 cache, signed version gate, and common-object prefix;
//   - one modifier-direction read and four low-byte name transfers;
//   - the live write-slot reload after each name-length callback;
//   - read-side modifier application before the live class/update/position
//     accesses used to reset a team flag's home position;
//   - a live Field34 reload, exact-one inventory gate, and success-only
//     Field34 restoration.
//
// There are intentionally no nil guards. The original faults at the first
// dereference reached in an invalid object-owned record. Stream callback
// return values are ignored just as they were by the original routine.
func teamXfer4F6D20[O, I any, S comparable, M, U any](
	object O,
	deps teamXferDeps4F6D20[O, I, S, M, U],
) int32 {
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(teamXferCurrentVersion4F6D20)
	version := int16(versionWord)
	if version > int16(teamXferCurrentVersion4F6D20) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readMode() != 0 {
		var modifiers [4]M
		for i := range modifiers {
			size := deps.rwByte(0)
			// GAME.EXE zero-extends the byte before comparing it with 256,
			// so every possible serialized byte length is accepted.
			name := deps.readModifierName(size)
			id := deps.modifierIDByName(name)
			modifiers[i] = deps.modifierByID(id)
		}
		deps.applyModifiers(object, modifiers, ^uint32(0))

		if deps.loadObjClass(object)&0x10000000 != 0 {
			update := deps.loadUpdateData(object)
			x := deps.loadPositionX(object)
			deps.storeUpdatePositionX(update, x)
			y := deps.loadPositionY(object)
			deps.storeUpdatePositionY(update, y)
		}
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

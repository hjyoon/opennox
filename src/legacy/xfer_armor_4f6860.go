package legacy

const armorXferCurrentVersion4F6860 = uint16(62)

// armorXferDeps4F6860 exposes every observable load, store, stream transfer,
// mode read, and callback in GAME.EXE 004F6860. Pointer-bearing records remain
// generic so the contract cannot inherit the original PE32 pointer width.
type armorXferDeps4F6860[
	O, I any,
	S comparable,
	M, H any,
	A comparable,
	W any,
] struct {
	loadField34  func(O) uint32
	storeField34 func(O, uint32)
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	readMode                  func() int32
	applyLegacyEmptyModifiers func(O)
	loadModifierData          func(O) I
	loadModifierSlot          func(I, int) S
	modifierNameLength        func(S) uint32
	rwByte                    func(uint8) uint8
	rwModifierName            func(S, uint8)
	readModifierName          func(uint8) string
	modifierIDByName          func(string) int32
	modifierByID              func(int32) M
	applyModifiers            func(O, [4]M, uint32)

	unitGetHP          func(O) uint16
	rwWord             func(uint16) uint16
	loadHealthData     func(O) H
	loadHealthMaximum  func(H) uint16
	switchToSolo       func() int32
	notMultiplayer     func() int32
	gameFlag4096       func() bool
	anyTrackedPlayers  func() int32
	unitSetHP          func(O, uint16)
	loadTypeIndex      func(O) uint16
	armorClass         func(uint16) A
	loadArmorHP        func(A) uint16
	storeHealthMaximum func(H, uint16)
	storeHealthField2  func(H, uint16)

	rwDummyByte       func()
	loadUpdateData    func(O) W
	rwUpdateField4    func(W)
	transferInventory func(uint16, O, int32) int32
}

// armorXfer4F6860 preserves the original PE32 transfer semantics while
// keeping all object identities native-width. In particular it preserves:
//
//   - the entry Field34 cache, signed version gate, and sign-extended common version;
//   - one cached modifier-mode read and the exact-one pre-v11 shortcut;
//   - four low-byte modifier-name transfers with live write-slot reloads;
//   - the HP clamp, exact-one apply gate, short-circuit predicates, and live fallback loads;
//   - the version-61 dummy byte, version-62 update field, and failure prefixes;
//   - Field34 restoration only after the complete suffix succeeds.
//
// There are intentionally no nil guards. The original faults at the first
// dereference reached in an invalid object-owned record. Stream callback
// return values are likewise ignored by the original routine.
func armorXfer4F6860[
	O, I any,
	S comparable,
	M, H any,
	A comparable,
	W any,
](object O, deps armorXferDeps4F6860[O, I, S, M, H, A, W]) int32 {
	originalField34 := deps.loadField34(object)
	versionWord := deps.rwVersion(armorXferCurrentVersion4F6860)
	version := int16(versionWord)
	if version > int16(armorXferCurrentVersion4F6860) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	// 004F68C8 loads the mode once and reuses that value for both the legacy
	// shortcut and the normal modifier direction.
	modifierMode := deps.readMode()
	if version < 11 && modifierMode == 1 {
		deps.applyLegacyEmptyModifiers(object)
		return 1
	}

	if modifierMode != 0 {
		var modifiers [4]M
		for i := range modifiers {
			size := deps.rwByte(0)
			// GAME.EXE compares the zero-extended byte with 256. That failure
			// edge is unreachable, so every possible byte length is accepted.
			name := deps.readModifierName(size)
			id := deps.modifierIDByName(name)
			modifiers[i] = deps.modifierByID(id)
		}
		deps.applyModifiers(object, modifiers, ^uint32(0))
	} else {
		data := deps.loadModifierData(object)
		var zero S
		for i := 0; i < 4; i++ {
			slot := deps.loadModifierSlot(data, i)
			if slot == zero {
				deps.rwByte(0)
				continue
			}
			size := deps.rwByte(uint8(deps.modifierNameLength(slot)))
			// The descriptor/name is loaded again after the length callback.
			deps.rwModifierName(deps.loadModifierSlot(data, i), size)
		}
	}

	if version >= 41 {
		hp := deps.rwWord(deps.unitGetHP(object))
		health := deps.loadHealthData(object)
		if maximum := deps.loadHealthMaximum(health); hp > maximum {
			hp = maximum
		}

		if deps.readMode() == 1 {
			direct := deps.switchToSolo() == 1
			if !direct {
				direct = deps.notMultiplayer() == 1
			}
			if !direct && deps.gameFlag4096() {
				direct = deps.anyTrackedPlayers() != 0
			}

			if direct {
				deps.unitSetHP(object, hp)
			} else {
				armor := deps.armorClass(deps.loadTypeIndex(object))
				var zero A
				if armor != zero {
					health := deps.loadHealthData(object)
					maximum := deps.loadArmorHP(armor)
					deps.storeHealthMaximum(health, maximum)

					health = deps.loadHealthData(object)
					field2 := deps.loadArmorHP(armor)
					deps.storeHealthField2(health, field2)

					deps.unitSetHP(object, deps.loadArmorHP(armor))
				}
			}
		}
	}

	if versionWord == 61 {
		// The original writes an uninitialized stack byte in save mode, so
		// only the one-byte stream event—not a stable input value—is contractual.
		deps.rwDummyByte()
	}
	if version >= 62 {
		deps.rwUpdateField4(deps.loadUpdateData(object))
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

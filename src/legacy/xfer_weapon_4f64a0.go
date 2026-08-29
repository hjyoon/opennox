package legacy

const (
	weaponXferCurrentVersion4F64A0 = uint16(64)

	weaponXferClassMask4F64A0      = uint32(0x00001000)
	weaponXferSubclassMask4F64A0   = uint32(0x047f0000)
	weaponXferLegacySkipMask4F64A0 = uint32(0x04000000)
)

// weaponXferDeps4F64A0 exposes every observable load, store, stream transfer,
// mode read, and callback in GAME.EXE 004F64A0. All pointer-bearing records
// remain generic so the contract cannot inherit PE32 pointer truncation.
type weaponXferDeps4F64A0[
	O, I any,
	S comparable,
	M, U, H any,
	P comparable,
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

	loadClass          func(O) uint32
	loadSubclass       func(O) uint32
	loadUseData        func(O) U
	loadChargeCurrent  func(U) uint8
	loadChargeMaximum  func(U) uint8
	loadChargeValue    func(U) int32
	rwDword            func(int32) int32
	storeChargeCurrent func(U, uint8)
	storeChargeMaximum func(U, uint8)
	storeChargeValue   func(U, int32)
	gameFlag4096       func() bool

	unitGetHP          func(O) uint16
	rwWord             func(uint16) uint16
	loadHealthData     func(O) H
	loadHealthMaximum  func(H) uint16
	switchToSolo       func() int32
	notMultiplayer     func() int32
	anyTrackedPlayers  func() int32
	unitSetHP          func(O, uint16)
	loadTypeIndex      func(O) uint16
	projectileClass    func(uint16) P
	loadProjectileHP   func(P) uint16
	storeHealthMaximum func(H, uint16)
	storeHealthField2  func(H, uint16)

	loadUpdateData    func(O) W
	rwUpdateField4    func(W)
	transferInventory func(uint16, O, int32) int32
}

// weaponXfer4F64A0 preserves the original PE32 transfer semantics while
// keeping all object identities native-width. In particular it preserves:
//
//   - the signed version gate and zero-extended inventory version;
//   - the single cached modifier-mode read and exact-one legacy shortcut;
//   - four low-byte modifier-name transfers with live write-slot reloads;
//   - the version-62 subclass exception and cached charge validation values;
//   - the HP clamp, short-circuit mode gates, and live fallback reloads;
//   - the version-63 dummy byte, version-64 update field, and failure prefixes;
//   - Field34 restoration only after the complete suffix succeeds.
//
// There are intentionally no nil guards. The original faults at the first
// dereference that an invalid object-owned record reaches. Stream callback
// return values are likewise ignored by the original routine.
func weaponXfer4F64A0[
	O, I any,
	S comparable,
	M, U, H any,
	P comparable,
	W any,
](object O, deps weaponXferDeps4F64A0[O, I, S, M, U, H, P, W]) int32 {
	originalField34 := deps.loadField34(object)
	versionWord := deps.rwVersion(weaponXferCurrentVersion4F64A0)
	version := int16(versionWord)
	if version > int16(weaponXferCurrentVersion4F64A0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	// 004F650B reads this mode once and reuses it for both the pre-v11
	// shortcut and the normal modifier direction.
	modifierMode := deps.readMode()
	if version < 11 && modifierMode == 1 {
		// Only the four PE32 modifier slots are cleared here. The fifth dword
		// in the stack record is not initialized, so this remains a distinct
		// hook rather than fabricating a stable Field16 value in Go.
		deps.applyLegacyEmptyModifiers(object)
		return 1
	}

	if modifierMode != 0 {
		var modifiers [4]M
		for i := range modifiers {
			size := deps.rwByte(0)
			// The original compares this byte with 256; the rejection edge is
			// unreachable after zero extension, so every byte value is accepted.
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
			// GAME.EXE reloads both the slot and its descriptor/name after the
			// length transfer instead of reusing the strlen operand.
			deps.rwModifierName(deps.loadModifierSlot(data, i), size)
		}
	}

	// v14 aliases the version-63 dummy-byte stack slot. When the charge
	// block executes it contains the transferred current charge. Otherwise
	// its original value is undefined; zero is a safe deterministic input for
	// write-mode compatibility, while read mode overwrites it.
	var scratchByte uint8
	if version >= 41 {
		classMask := deps.loadClass(object) & weaponXferClassMask4F64A0
		eligible := false
		if classMask != 0 {
			eligible = deps.loadSubclass(object)&weaponXferSubclassMask4F64A0 != 0
		}

		skipLegacyCharge := false
		if version < 62 && classMask != 0 {
			// This is a second, live subclass load even when the first load did
			// not make the object eligible.
			skipLegacyCharge = deps.loadSubclass(object)&weaponXferLegacySkipMask4F64A0 != 0
		}
		if eligible && !skipLegacyCharge {
			use := deps.loadUseData(object)
			oldCurrent := deps.loadChargeCurrent(use)
			oldMaximum := deps.loadChargeMaximum(use)
			oldValue := deps.loadChargeValue(use)

			_ = oldCurrent // cached by the original even though it is not validated
			scratchByte = deps.rwByte(deps.loadChargeCurrent(use))
			newMaximum := deps.rwByte(deps.loadChargeMaximum(use))
			newValue := oldValue
			if version >= 61 {
				newValue = deps.rwDword(deps.loadChargeValue(use))
			}

			valid := scratchByte <= oldMaximum &&
				newValue >= 0 && newValue <= oldValue &&
				newMaximum == oldMaximum
			if !deps.gameFlag4096() || valid {
				deps.storeChargeCurrent(use, scratchByte)
				deps.storeChargeMaximum(use, newMaximum)
				deps.storeChargeValue(use, newValue)
			} else {
				deps.storeChargeCurrent(use, 0)
				deps.storeChargeMaximum(use, oldMaximum)
				deps.storeChargeValue(use, 0)
			}
		}
	}

	if version >= 42 {
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
				projectile := deps.projectileClass(deps.loadTypeIndex(object))
				var zero P
				if projectile != zero {
					health := deps.loadHealthData(object)
					maximum := deps.loadProjectileHP(projectile)
					deps.storeHealthMaximum(health, maximum)

					health = deps.loadHealthData(object)
					field2 := deps.loadProjectileHP(projectile)
					deps.storeHealthField2(health, field2)

					deps.unitSetHP(object, deps.loadProjectileHP(projectile))
				}
			}
		}
	}

	if versionWord == 63 {
		scratchByte = deps.rwByte(scratchByte)
	}
	if version >= 64 {
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

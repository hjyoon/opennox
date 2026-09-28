package server

const (
	weaponCreateQuestFlag54C710 = uint32(0x00001000)

	weaponCreateAmmoClass54C710         = uint32(0x01000000)
	weaponCreateStaffClass54C710        = uint32(0x00001000)
	weaponCreateAmmoChargeMask54C710    = uint32(0x00000082)
	weaponCreateAmmoEmptyMask54C710     = uint32(0x0000000c)
	weaponCreateStaffMask54C710         = uint32(0x047f0000)
	weaponCreateDoubleChargeMask54C710  = uint32(0x00040000)
	weaponCreateAmmoCharge054C710       = uintptr(0)
	weaponCreateAmmoCharge154C710       = uintptr(1)
	weaponCreateAmmoField254C710        = uintptr(2)
	weaponCreateWandCharge54C710        = uintptr(108)
	weaponCreateWandMaxCharge54C710     = uintptr(109)
	weaponCreateHeartType54C710         = "OblivionHeart"
	weaponCreateWierdlingType54C710     = "OblivionWierdling"
	weaponCreateOrbType54C710           = "OblivionOrb"
	weaponCreateHeartModifier54C710     = "Lightning4"
	weaponCreateWierdlingModifier54C710 = "Vampirism2"
	weaponCreateWierdlingSecond54C710   = "Lightning3"
	weaponCreateDurabilityKey54C710     = "QuestDurabilityMultiplier"
	weaponCreateAmmoQuestKey54C710      = "DefaultAmmoAmountQuest"
	weaponCreateAmmoDefaultKey54C710    = "DefaultAmmoAmount"
	weaponCreateStaffChargeKey54C710    = "QuestStaffChargeMultiplier"
)

type weaponCreateTypeCache54C710 struct {
	heart     uint32
	wierdling uint32
	orb       uint32
}

type weaponCreateHooks54C710[O, D, H, A, E, U any] struct {
	loadTypeInd    func(O) uint16
	loadInitData   func(O) A
	findDefinition func(uint16) D

	loadHeartType      func() uint32
	storeHeartType     func(uint32)
	loadWierdlingType  func() uint32
	storeWierdlingType func(uint32)
	storeOrbType       func(uint32)
	resolveObjectType  func(string) uint32

	loadHealth     func(O) H
	loadDurability func(D) uint16
	storeCurrent   func(H, uint16)
	storeMaximum   func(H, uint16)
	loadCurrent    func(H) uint16
	loadMaximum    func(H) uint16

	modifierID    func(string) int
	modifierDesc  func(int) E
	storeModifier func(A, int, E)

	loadClass    func(O) uint32
	loadSubclass func(O) uint32
	loadUseData  func(O) U
	loadUseByte  func(U, uintptr) uint8
	storeUseByte func(U, uintptr, uint8)

	gameFlag    func(uint32) int32
	loadBalance func(string) float32
	floatToInt  func(float32) int32
}

// weaponCreate54C710 preserves GAME.EXE 0054C710. InitData and the weapon
// definition are cached at entry, while HealthData, class, subclass, and
// UseData retain the executable's individual live-load points. In particular,
// the special Oblivion modifier and ammo/staff stores intentionally have no
// nil guard: those guards appeared in the recovered C but not in the original
// PE32 instruction stream.
func weaponCreate54C710[O any, D, H comparable, A, E, U any](obj O, hooks weaponCreateHooks54C710[O, D, H, A, E, U]) {
	typeInd := hooks.loadTypeInd(obj)
	initData := hooks.loadInitData(obj)
	definition := hooks.findDefinition(typeInd)

	heartType := hooks.loadHeartType()
	if heartType == 0 {
		heartType = hooks.resolveObjectType(weaponCreateHeartType54C710)
		hooks.storeHeartType(heartType)
		wierdlingType := hooks.resolveObjectType(weaponCreateWierdlingType54C710)
		hooks.storeWierdlingType(wierdlingType)
		orbType := hooks.resolveObjectType(weaponCreateOrbType54C710)
		hooks.storeOrbType(orbType)
	}

	var nilDefinition D
	if definition != nilDefinition {
		health := hooks.loadHealth(obj)
		var nilHealth H
		if health != nilHealth {
			durability := hooks.loadDurability(definition)
			hooks.storeCurrent(health, durability)
			health = hooks.loadHealth(obj)
			durability = hooks.loadDurability(definition)
			hooks.storeMaximum(health, durability)

			if hooks.gameFlag(weaponCreateQuestFlag54C710) != 0 {
				multiplier := hooks.loadBalance(weaponCreateDurabilityKey54C710)
				health = hooks.loadHealth(obj)
				current := hooks.loadCurrent(health)
				currentValue := hooks.floatToInt(weaponCreateScale54C710(current, multiplier))
				health = hooks.loadHealth(obj)
				hooks.storeCurrent(health, uint16(currentValue))

				health = hooks.loadHealth(obj)
				maximum := hooks.loadMaximum(health)
				maximumValue := hooks.floatToInt(weaponCreateScale54C710(maximum, multiplier))
				health = hooks.loadHealth(obj)
				hooks.storeMaximum(health, uint16(maximumValue))
			}
		}
	}

	heartType = hooks.loadHeartType()
	typeInd = hooks.loadTypeInd(obj)
	if uint32(typeInd) == heartType {
		modifierID := hooks.modifierID(weaponCreateHeartModifier54C710)
		modifier := hooks.modifierDesc(modifierID)
		hooks.storeModifier(initData, 2, modifier)
	} else if uint32(typeInd) == hooks.loadWierdlingType() {
		modifierID := hooks.modifierID(weaponCreateWierdlingModifier54C710)
		modifier := hooks.modifierDesc(modifierID)
		hooks.storeModifier(initData, 2, modifier)
		modifierID = hooks.modifierID(weaponCreateWierdlingSecond54C710)
		modifier = hooks.modifierDesc(modifierID)
		hooks.storeModifier(initData, 3, modifier)
	}

	if hooks.loadClass(obj)&weaponCreateAmmoClass54C710 != 0 {
		subclass := hooks.loadSubclass(obj)
		if subclass&weaponCreateAmmoChargeMask54C710 != 0 {
			useData := hooks.loadUseData(obj)
			balanceKey := weaponCreateAmmoDefaultKey54C710
			if hooks.gameFlag(weaponCreateQuestFlag54C710) != 0 {
				balanceKey = weaponCreateAmmoQuestKey54C710
			}
			amount := hooks.loadBalance(balanceKey)
			value := uint8(hooks.floatToInt(amount))
			hooks.storeUseByte(useData, weaponCreateAmmoCharge154C710, value)
			hooks.storeUseByte(useData, weaponCreateAmmoField254C710, 0)
			hooks.storeUseByte(useData, weaponCreateAmmoCharge054C710, value)
		} else if subclass&weaponCreateAmmoEmptyMask54C710 != 0 {
			useData := hooks.loadUseData(obj)
			hooks.storeUseByte(useData, weaponCreateAmmoCharge054C710, 0)
		}
	}

	if hooks.gameFlag(weaponCreateQuestFlag54C710) == 0 {
		return
	}
	if hooks.loadClass(obj)&weaponCreateStaffClass54C710 == 0 {
		return
	}
	if hooks.loadSubclass(obj)&weaponCreateStaffMask54C710 == 0 {
		return
	}

	useData := hooks.loadUseData(obj)
	multiplier := hooks.loadBalance(weaponCreateStaffChargeKey54C710)
	if hooks.loadSubclass(obj)&weaponCreateDoubleChargeMask54C710 != 0 {
		multiplier = weaponCreateDoubleMultiplier54C710(multiplier)
	}
	maximum := hooks.loadUseByte(useData, weaponCreateWandMaxCharge54C710)
	maximumValue := hooks.floatToInt(weaponCreateScaleByte54C710(maximum, multiplier))
	hooks.storeUseByte(useData, weaponCreateWandMaxCharge54C710, uint8(maximumValue))
	current := hooks.loadUseByte(useData, weaponCreateWandCharge54C710)
	currentValue := hooks.floatToInt(weaponCreateScaleByte54C710(current, multiplier))
	hooks.storeUseByte(useData, weaponCreateWandCharge54C710, uint8(currentValue))
}

// weaponCreateScale54C710 models FILD/FMULS followed by the executable's
// binary32 spill before nox_float2int.
func weaponCreateScale54C710(value uint16, multiplier float32) float32 {
	return float32(float64(value) * float64(multiplier))
}

func weaponCreateScaleByte54C710(value uint8, multiplier float32) float32 {
	return float32(float64(value) * float64(multiplier))
}

// weaponCreateDoubleMultiplier54C710 models FLD/FADD ST,ST/FSTP for the
// lightning-staff Quest charge special case.
func weaponCreateDoubleMultiplier54C710(multiplier float32) float32 {
	return float32(float64(multiplier) + float64(multiplier))
}

package server

import "math"

const (
	poisonProtectionEquippedFlag4E0040 = uint32(0x00000100)
	poisonProtectionClassMask4E0040    = uint32(0x13001000)
	poisonProtectionEnchant4E0040      = uint32(18)
	poisonProtectionModifierSlots      = 4
	poisonProtectionModifierLimitBits  = uint32(0x3f333333)
	poisonProtectionFinalLimitBits     = uint32(0x3f666666)
	poisonProtectionBalanceKey4E0040   = "PoisonSpellProtection"
)

type poisonProtectionHooks4E0040[O, I, M any] struct {
	loadUnitArg       func() O
	loadFirstItem     func(O) O
	loadFlags         func(O) uint32
	loadClass         func(O) uint32
	loadInitData      func(O) I
	loadModifier      func(I, int) M
	matchesProtection func(M) bool
	loadModifierValue func(M) float32
	loadNextItem      func(O) O
	testBuff          func(O, uint32) int32
	loadBuffPower     func(O, uint32) uint32
	loadBalance       func(string, int32) float64
}

// poisonProtection4E0040 preserves GAME.EXE 004E0040. Inventory modifier
// values and unit modifier values share one x87 accumulator. Its final value
// is spilled once to binary32 before the modifier-only clamp. Optional spell
// protection is then added at 53-bit precision before the final clamp.
func poisonProtection4E0040[O, I, M comparable](hooks poisonProtectionHooks4E0040[O, I, M]) float64 {
	unit := hooks.loadUnitArg()
	var nilObject O
	if unit == nilObject {
		return 0
	}

	accumulator := float64(0)
	subtotal := float32(0)
	if item := hooks.loadFirstItem(unit); item != nilObject {
		for item != nilObject {
			flags := hooks.loadFlags(item)
			if flags&poisonProtectionEquippedFlag4E0040 != 0 {
				class := hooks.loadClass(item)
				if class&poisonProtectionClassMask4E0040 != 0 {
					accumulator = poisonProtectionAddModifiers4E0040(hooks, accumulator, hooks.loadInitData(item))
				}
			}
			item = hooks.loadNextItem(item)
		}
		subtotal = float32(accumulator)
	}

	class := hooks.loadClass(unit)
	if class&poisonProtectionClassMask4E0040 != 0 {
		accumulator = poisonProtectionAddModifiers4E0040(hooks, accumulator, hooks.loadInitData(unit))
		subtotal = float32(accumulator)
	}

	modifierLimit := math.Float32frombits(poisonProtectionModifierLimitBits)
	if accumulator > float64(modifierLimit) {
		subtotal = modifierLimit
	}

	result := float64(subtotal)
	if hooks.testBuff(unit, poisonProtectionEnchant4E0040) != 0 {
		power := uint8(hooks.loadBuffPower(unit, poisonProtectionEnchant4E0040))
		index := int32(uint32(power) - 1)
		result = hooks.loadBalance(poisonProtectionBalanceKey4E0040, index) + float64(subtotal)
	}

	finalLimit := math.Float32frombits(poisonProtectionFinalLimitBits)
	if result > float64(finalLimit) {
		return float64(finalLimit)
	}
	return result
}

func poisonProtectionAddModifiers4E0040[O, I, M comparable](
	hooks poisonProtectionHooks4E0040[O, I, M],
	accumulator float64,
	initData I,
) float64 {
	var nilModifier M
	for slot := 0; slot < poisonProtectionModifierSlots; slot++ {
		modifier := hooks.loadModifier(initData, slot)
		if modifier != nilModifier && hooks.matchesProtection(modifier) {
			accumulator += float64(hooks.loadModifierValue(modifier))
		}
	}
	return accumulator
}

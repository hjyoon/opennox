package server

import (
	"math"
	"unsafe"
)

const (
	fireProtectionEquippedFlag4DFE40 = uint32(0x00000100)
	fireProtectionClassMask4DFE40    = uint32(0x13001000)
	fireProtectionModifierLimitBits  = uint32(0x3f000000)
	fireProtectionFinalLimitBits     = uint32(0x3f19999a)
	fireProtectionBalanceKey4DFE40   = "FireSpellProtection"
)

// FireProtectionRuntime4DFE40 carries the identity of the original
// FireProtectEngage callback. GAME.EXE compares this identity without calling
// the callback while it sums fire protection.
type FireProtectionRuntime4DFE40 struct {
	FireProtectEngage unsafe.Pointer
}

func fireProtectionAddModifiers4DFE40(init *ModifierInitData, engage unsafe.Pointer, accumulator float64) float64 {
	for _, modifier := range init.Modifiers {
		if modifier != nil && engage != nil && modifier.Engage112 == engage {
			accumulator += float64(modifier.EngageFloat120)
		}
	}
	return accumulator
}

func fireProtectionNative4DFE40(
	unit *Object,
	engage unsafe.Pointer,
	loadBalance func(string, int32) float64,
) float64 {
	if unit == nil {
		return 0
	}

	accumulator := float64(0)
	subtotal := float32(0)
	if item := unit.InvFirstItem; item != nil {
		for ; item != nil; item = item.InvNextItem {
			if uint32(item.ObjFlags)&fireProtectionEquippedFlag4DFE40 != 0 &&
				uint32(item.ObjClass)&fireProtectionClassMask4DFE40 != 0 {
				accumulator = fireProtectionAddModifiers4DFE40(item.InitDataModifier(), engage, accumulator)
			}
		}
		subtotal = float32(accumulator)
	}
	if uint32(unit.ObjClass)&fireProtectionClassMask4DFE40 != 0 {
		accumulator = fireProtectionAddModifiers4DFE40(unit.InitDataModifier(), engage, accumulator)
		subtotal = float32(accumulator)
	}
	modifierLimit := math.Float32frombits(fireProtectionModifierLimitBits)
	if accumulator > float64(modifierLimit) {
		subtotal = modifierLimit
	}

	result := float64(subtotal)
	if unit.HasEnchant(ENCHANT_PROTECT_FROM_FIRE) {
		power := uint8(unit.EnchantPower(ENCHANT_PROTECT_FROM_FIRE))
		index := int32(uint32(power) - 1)
		result = loadBalance(fireProtectionBalanceKey4DFE40, index) + float64(subtotal)
	}
	finalLimit := math.Float32frombits(fireProtectionFinalLimitBits)
	if result > float64(finalLimit) {
		return float64(finalLimit)
	}
	return result
}

// FireProtection4DFE40 binds GAME.EXE 004DFE40 to native-width Object and
// modifier layouts.
func (s *Server) FireProtection4DFE40(unit *Object, runtime FireProtectionRuntime4DFE40) float64 {
	return fireProtectionNative4DFE40(unit, runtime.FireProtectEngage, func(key string, index int32) float64 {
		return s.Balance.FloatInd(key, int(index))
	})
}

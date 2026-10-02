package server

const (
	shockGlyphName52C5A0      = "Glyph"
	shockTrapDamageKey52C5A0  = "ShockTrapDamage"
	shockDurationKey52C5A0    = "ShockEnchantDuration"
	shockBuff52C5A0           = int32(22)
	shockElectricDamage52C5A0 = int32(9)
)

type castShockHooks52C5A0[O comparable] struct {
	target         func() O
	loadGlyphType  func() uint32
	storeGlyphType func(uint32)
	lookupType     func(string) uint32
	typeIndex      func(O) uint16
	balance        func(string) float64
	balanceIndex   func(string, int32) float64
	floatToInt     func(float32) int32
	damage         func(O, O, O, int32, int32) bool
	apply          func(O, int32, int16, int8)
}

// castShock52C5A0 preserves GAME.EXE 0052C5A0..0052C63E. Only an empty
// acceptance target returns zero. The whole-DWORD Glyph cache is initialized
// before testing the fourth argument's zero-extended type word. A matching
// Glyph calls the live target's damage handler with the third argument twice;
// other contexts apply SHOCK. Both balance values spill to binary32 before
// nearest-even conversion, and both branches reload the target afterwards.
// Damage rejection does not change the canonical success result.
func castShock52C5A0[O comparable](caster, context O, power int32, h castShockHooks52C5A0[O]) int32 {
	var nilObject O
	if h.target() == nilObject {
		return 0
	}
	glyph := h.loadGlyphType()
	if glyph == 0 {
		glyph = h.lookupType(shockGlyphName52C5A0)
		h.storeGlyphType(glyph)
	}
	if context != nilObject && uint32(h.typeIndex(context)) == glyph {
		damage := h.floatToInt(float32(h.balanceIndex(shockTrapDamageKey52C5A0, power-1)))
		h.damage(h.target(), caster, caster, damage, shockElectricDamage52C5A0)
	} else {
		duration := h.floatToInt(float32(h.balance(shockDurationKey52C5A0)))
		h.apply(h.target(), shockBuff52C5A0, int16(duration), int8(power))
	}
	return 1
}

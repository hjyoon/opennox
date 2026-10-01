package server

const stunDurationKey52C2C0 = "StunEnchantDuration"

type castStunHooks52C2C0[O comparable] struct {
	target      func() O
	balance     func(string) float64
	floatToInt  func(float32) int32
	classLow    func(O) uint8
	playerClass func(O) uint8
	mass        func(O) float32
	apply       func(O, int32, int16, int8)
	attribution func(O, O)
}

// castStun52C2C0 preserves GAME.EXE 0052C2C0. Players take precedence
// over monsters: the raw zero class selects SLOWED, all nonzero classes
// select HELD. Only non-player monsters read Mass (PE32 Object+120);
// strictly greater than 15 selects SLOWED; unordered comparisons select HELD.
// The target is cached for these reads and application, then reloaded for
// attribution. Duration spills to binary32 before 00419A70 rounds it.
func castStun52C2C0[O comparable](caster O, power int32, h castStunHooks52C2C0[O]) int32 {
	var nilObject O
	if h.target() == nilObject {
		return 0
	}
	duration := h.floatToInt(float32(h.balance(stunDurationKey52C2C0)))
	target := h.target()
	class := h.classLow(target)
	buff := int32(5)
	if class&4 != 0 {
		if h.playerClass(target) == 0 {
			buff = 4
		}
	} else if class&2 != 0 && h.mass(target) > 15 {
		buff = 4
	}
	h.apply(target, buff, int16(duration), int8(power))
	h.attribution(caster, h.target())
	return 1
}

package server

const confuseDurationKey52C1E0 = "ConfuseEnchantDuration"

type castConfuseHooks52C1E0[O comparable] struct {
	target      func() O
	balance     func(string) float64
	floatToInt  func(float32) int32
	apply       func(O, int32, int16, int8)
	attribution func(O, O)
}

// castConfuse52C1E0 preserves GAME.EXE 0052C1E0, including the live target
// reloads around balance lookup and buff application. The duration is spilled
// to binary32 before 00419A70 rounds it; only the buff callee narrows its word
// duration and byte power. A rejected buff still returns success here.
func castConfuse52C1E0[O comparable](caster O, power int32, h castConfuseHooks52C1E0[O]) int32 {
	var nilObject O
	if h.target() == nilObject {
		return 0
	}
	duration := h.floatToInt(float32(h.balance(confuseDurationKey52C1E0)))
	h.apply(h.target(), 3, int16(duration), int8(power))
	h.attribution(caster, h.target())
	return 1
}

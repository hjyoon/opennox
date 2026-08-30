package server

const (
	playerConfusedBuff4F7A40        = uint32(3)
	playerConfusedPeriod4F7A40      = uint32(40)
	playerConfusedHalfPeriod4F7A40  = uint32(20)
	playerConfusedPhaseCenter4F7A40 = int32(10)
	playerConfusedPowerBias4F7A40   = int32(3)
)

type playerConfusedDirectionHooks4F7A40[O any] struct {
	loadDirection2 func(O) uint16
	loadBuffPower  func(O, uint32) uint8
	loadFrame      func() uint32
	loadNetCode    func(O) uint32
}

// playerConfusedDirection4F7A40 preserves GAME.EXE 004F7A40. Direction2 is
// cached as a signed 16-bit value before the buff-power callback. Frame and
// NetCode are then loaded in that order, so callback-side mutations remain
// visible. The unsigned frame sum wraps before the 40-frame triangular wave.
func playerConfusedDirection4F7A40[O any](
	unit O,
	hooks playerConfusedDirectionHooks4F7A40[O],
) uint32 {
	direction := int32(int16(hooks.loadDirection2(unit)))
	power := int32(hooks.loadBuffPower(unit, playerConfusedBuff4F7A40))
	frame := hooks.loadFrame()
	netCode := hooks.loadNetCode(unit)

	phase := (frame + netCode) % playerConfusedPeriod4F7A40
	if phase > playerConfusedHalfPeriod4F7A40 {
		phase = playerConfusedPeriod4F7A40 - phase
	}
	value := direction + (power+playerConfusedPowerBias4F7A40)*
		(int32(phase)-playerConfusedPhaseCenter4F7A40)

	// Mirror the two IA-32 normalization branches instead of relying on a
	// signed remainder: the public result is the canonical range 0..255.
	if value < 0 {
		value += int32((uint32(255-value) >> 8) << 8)
	}
	if value >= 256 {
		value -= 256 * int32(uint32(value)>>8)
	}
	return uint32(value)
}

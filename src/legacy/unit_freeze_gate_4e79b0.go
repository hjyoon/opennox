package legacy

// unitFreezeGateSet4E79B0 is the fixed-width contract for GAME.EXE
// 004E79B0. The original stores the argument once and returns the same 32
// bits; it does not read the previous gate value.
func unitFreezeGateSet4E79B0(value uint32, store func(uint32)) uint32 {
	store(value)
	return value
}

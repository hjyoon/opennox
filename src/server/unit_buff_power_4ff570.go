package server

// UnitBuffPowerHooks4FF570 exposes the three observable loads in GAME.EXE
// 004FF570. Object tokens stay native-width, while the buff argument and
// power retain their original dword and byte widths.
type UnitBuffPowerHooks4FF570[Object any] struct {
	LoadBuffArg func() int32
	LoadUnitArg func() Object
	LoadPower   func(Object, int32) uint8
}

// UnitBuffPower4FF570 preserves GAME.EXE 004FF570's argument-load order,
// unchecked full-dword index, and one-byte result. The original reads buff
// before unit and has no null or range guard. LoadPower models the PE32
// effective-address read at unit+0x198+buff. The IA-32 body returns a C char
// through AL; the retained upper EAX bytes are outside that result contract.
func UnitBuffPower4FF570[Object any](h UnitBuffPowerHooks4FF570[Object]) uint8 {
	buff := h.LoadBuffArg()
	unit := h.LoadUnitArg()
	return h.LoadPower(unit, buff)
}

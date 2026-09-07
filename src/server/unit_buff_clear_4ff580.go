package server

const unitBuffClearSlots4FF580 = int32(32)

// UnitBuffClearHooks4FF580 exposes the argument load, flag callback, and exact
// word/byte stores performed by GAME.EXE 004FF580. Object tokens stay
// native-width while slot indices retain the original dword width.
type UnitBuffClearHooks4FF580[Object any] struct {
	LoadUnitArg   func() Object
	SetBuffFlags  func(Object, uint32)
	StoreDuration func(Object, int32, uint16)
	StorePower    func(Object, int32, uint8)
}

// UnitBuffClear4FF580 preserves GAME.EXE 004FF580's exact callback and store
// order. The unit argument is loaded once and retained across the flag
// callback. Each of the 32 duration words is cleared before its matching power
// byte. The original has no null or callback guard; native bindings therefore
// retain the same fail-fast boundary. Its final EAX value is 32, but the public
// ABI is void and every decoded caller discards that register value.
func UnitBuffClear4FF580[Object any](h UnitBuffClearHooks4FF580[Object]) {
	unit := h.LoadUnitArg()
	h.SetBuffFlags(unit, 0)
	for buff := int32(0); buff < unitBuffClearSlots4FF580; buff++ {
		h.StoreDuration(unit, buff, 0)
		h.StorePower(unit, buff, 0)
	}
}

package server

// UnitBuffTimerHooks4FF550 exposes the three observable loads in GAME.EXE
// 004FF550. Object tokens stay native-width, while the buff argument and
// duration retain their original dword and word widths.
type UnitBuffTimerHooks4FF550[Object any] struct {
	LoadBuffArg  func() int32
	LoadUnitArg  func() Object
	LoadDuration func(Object, int32) uint16
}

// UnitBuffTimer4FF550 preserves GAME.EXE 004FF550's argument-load order,
// unchecked full-dword index, and zero-extended word result. The original
// reads buff before unit and has no null or range guard. LoadDuration models
// the PE32 effective-address read at unit+0x158+2*buff; native bindings may
// deliberately turn an out-of-range address into a memory-safety fault.
func UnitBuffTimer4FF550[Object any](h UnitBuffTimerHooks4FF550[Object]) uint32 {
	buff := h.LoadBuffArg()
	unit := h.LoadUnitArg()
	return uint32(h.LoadDuration(unit, buff))
}

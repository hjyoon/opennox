package server

type directionInitHooks4F0490[O, I any] struct {
	loadInitData     func(O) I
	directionToAngle func(I) uint32
	storeDirection2  func(O, uint16)
	storeDirection1  func(O, uint16)
}

// directionInit4F0490 preserves GAME.EXE 004F0490's observable order. The
// InitData pointer is loaded once before the helper call. The helper's low word
// is written to Direction2 before Direction1, while its full 32-bit result is
// returned unchanged. The original has no nil guards.
func directionInit4F0490[O, I any](unit O, hooks directionInitHooks4F0490[O, I]) int32 {
	initData := hooks.loadInitData(unit)
	angle := hooks.directionToAngle(initData)
	hooks.storeDirection2(unit, uint16(angle))
	hooks.storeDirection1(unit, uint16(angle))
	return int32(angle)
}

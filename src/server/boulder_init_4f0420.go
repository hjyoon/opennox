package server

type boulderInitHooks4F0420[O, V any] struct {
	loadSourceX  func(O) V
	loadSourceY  func(O) V
	storeTargetX func(O, V)
	storeTargetY func(O, V)
}

// boulderInit4F0420 preserves the exact observable order of GAME.EXE
// 004F0420. Both 32-bit coordinate values are loaded before either target is
// written, the X store precedes the Y store, and the entry object pointer is
// returned unchanged. The original has no nil guard.
func boulderInit4F0420[O, V any](unit O, hooks boulderInitHooks4F0420[O, V]) O {
	x := hooks.loadSourceX(unit)
	y := hooks.loadSourceY(unit)
	hooks.storeTargetX(unit, x)
	hooks.storeTargetY(unit, y)
	return unit
}

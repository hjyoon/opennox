package legacy

const scriptCallbackInitGameFlag4F5540 = uint32(0x400000)

type scriptCallbackInitDeps4F5540[H, F any] struct {
	readOnly      func() int32
	mapgenFile    func() F
	makeScript    func(F, H) int32
	gameFlagCheck func(uint32) int32
	storeFunc     func(H, int32)
}

// scriptCallbackInit4F5540 preserves GAME.EXE 004F5540. The global mode is
// loaded exactly once. Values other than one are returned without touching the
// handler or the remaining dependencies. In mode one, the script parser result
// is ignored, the game-flag result is returned verbatim, and only a zero flag
// result writes -1 to the handler's Func field. The original has no nil guard
// and does not roll back parser mutations when a later operation faults.
func scriptCallbackInit4F5540[H, F any](
	handler H,
	deps scriptCallbackInitDeps4F5540[H, F],
) int32 {
	result := deps.readOnly()
	if result != 1 {
		return result
	}
	file := deps.mapgenFile()
	_ = deps.makeScript(file, handler)
	result = deps.gameFlagCheck(scriptCallbackInitGameFlag4F5540)
	if result == 0 {
		deps.storeFunc(handler, -1)
	}
	return result
}

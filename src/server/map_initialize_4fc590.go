package server

const (
	mapInitializeScriptPrefix4FC590 = "MapInitialize"
	mapInitializeScriptStride4FC590 = uint32(48)
)

type mapInitializeHooks4FC590[T any] struct {
	loadState         func() int32
	hasPlayerUnit     func() bool
	loadScriptCount   func() int32
	loadScriptTable   func() T
	loadScriptName    func(T, uint32) string
	callScriptByIndex func(int32, uintptr, uintptr) int32
	clearState        func(int32) int32
}

// mapInitialize4FC590 preserves GAME.EXE 004FC590. Once the state and player
// gates pass, the original walks 48-byte legacy-script records. It reloads
// both the table base for every name and the signed count after every record,
// including across script callbacks, then clears the complete state dword.
func mapInitialize4FC590[T any](hooks mapInitializeHooks4FC590[T]) int32 {
	state := hooks.loadState()
	if state == 0 {
		return state
	}
	if !hooks.hasPlayerUnit() {
		return 0
	}

	index := int32(0)
	offset := uint32(0)
	if hooks.loadScriptCount() > 0 {
		for {
			table := hooks.loadScriptTable()
			name := hooks.loadScriptName(table, offset)
			if len(name) >= len(mapInitializeScriptPrefix4FC590) &&
				name[:len(mapInitializeScriptPrefix4FC590)] == mapInitializeScriptPrefix4FC590 {
				_ = hooks.callScriptByIndex(index, 0, 0)
			}

			index++
			offset += mapInitializeScriptStride4FC590
			if index >= hooks.loadScriptCount() {
				break
			}
		}
	}
	return hooks.clearState(0)
}

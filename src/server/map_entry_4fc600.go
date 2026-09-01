package server

const (
	mapEntryScriptPrefix4FC600 = "MapEntry"
	mapEntryScriptStride4FC600 = uint32(48)
)

type mapEntryHooks4FC600[T any] struct {
	loadState         func() int32
	hasPlayerUnit     func() bool
	loadScriptCount   func() int32
	loadScriptTable   func() T
	loadScriptName    func(T, uint32) string
	callScriptByIndex func(int32, uintptr, uintptr) int32
	clearState        func(int32) int32
}

// mapEntry4FC600 preserves GAME.EXE 004FC600. Once the state and player gates
// pass, the original walks 48-byte legacy-script records. It reloads the table
// base for every name and the signed count after every record, including across
// script callbacks, then clears the complete map-entry state dword.
func mapEntry4FC600[T any](hooks mapEntryHooks4FC600[T]) int32 {
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
			if len(name) >= len(mapEntryScriptPrefix4FC600) &&
				name[:len(mapEntryScriptPrefix4FC600)] == mapEntryScriptPrefix4FC600 {
				_ = hooks.callScriptByIndex(index, 0, 0)
			}

			index++
			offset += mapEntryScriptStride4FC600
			if index >= hooks.loadScriptCount() {
				break
			}
		}
	}
	return hooks.clearState(0)
}

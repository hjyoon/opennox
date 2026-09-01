package server

type mapEntryStateSetHooks4FC580 struct {
	loadValueArg func() int32
	storeValue   func(int32)
}

// mapEntryStateSet4FC580 preserves GAME.EXE 004FC580. The original cdecl
// routine reads the complete signed dword argument, stores that same bit
// pattern in the map-entry state, and returns it unchanged in EAX.
func mapEntryStateSet4FC580(hooks mapEntryStateSetHooks4FC580) int32 {
	value := hooks.loadValueArg()
	hooks.storeValue(value)
	return value
}

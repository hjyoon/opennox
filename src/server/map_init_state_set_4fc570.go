package server

type mapInitStateSetHooks4FC570 struct {
	loadValueArg func() int32
	storeValue   func(int32)
}

// mapInitStateSet4FC570 preserves GAME.EXE 004FC570. The original cdecl
// routine reads the complete signed dword argument, stores that same bit
// pattern in the map-initialize state, and returns it unchanged in EAX.
func mapInitStateSet4FC570(hooks mapInitStateSetHooks4FC570) int32 {
	value := hooks.loadValueArg()
	hooks.storeValue(value)
	return value
}

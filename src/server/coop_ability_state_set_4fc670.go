package server

type coopAbilityStateSetHooks4FC670 struct {
	loadValueArg func() int32
	storeValue   func(int32)
}

// coopAbilityStateSet4FC670 preserves GAME.EXE 004FC670. The original cdecl
// routine reads the complete signed dword argument, stores that same bit
// pattern in the queued cooperative-ability state, and returns it unchanged
// in EAX.
func coopAbilityStateSet4FC670(hooks coopAbilityStateSetHooks4FC670) int32 {
	value := hooks.loadValueArg()
	hooks.storeValue(value)
	return value
}

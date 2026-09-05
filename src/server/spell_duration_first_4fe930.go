package server

// SpellDurationFirstHooks4FE930 separates the first-record accessor contract
// from the native duration-spell list head.
type SpellDurationFirstHooks4FE930[R any] struct {
	LoadHead func() R
}

// SpellDurationFirst4FE930 preserves GAME.EXE 004FE930. The original PE32
// function performs exactly one live load of the duration-spell list head and
// returns that value unchanged, including nil. It does not validate the head,
// traverse a record, or narrow a native-width pointer.
func SpellDurationFirst4FE930[R any](hooks SpellDurationFirstHooks4FE930[R]) R {
	return hooks.LoadHead()
}

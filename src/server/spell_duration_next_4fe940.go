package server

// SpellDurationNextHooks4FE940 separates the next-record accessor contract
// from the native duration-spell record layout.
type SpellDurationNextHooks4FE940[R any] struct {
	LoadNext func(R) R
}

// SpellDurationNext4FE940 preserves GAME.EXE 004FE940. The original PE32
// function performs exactly one load of record->Next and returns that value
// unchanged, including nil. It does not validate the record, traverse beyond
// the loaded link, or narrow a native-width pointer.
func SpellDurationNext4FE940[R any](record R, hooks SpellDurationNextHooks4FE940[R]) R {
	return hooks.LoadNext(record)
}

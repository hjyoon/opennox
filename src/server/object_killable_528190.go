package server

// ObjectKillableNative528190 preserves the raw GAME.EXE 00528190 entry.
// A nil Object faults; missing health returns0. A zero maximum satisfies the
// predicate even at zero current health, as in the original.
func ObjectKillableNative528190(unit *Object) int32 {
	health := unit.HealthData
	if health == nil {
		return 0
	}
	current, maximum := health.Cur, health.Max
	if current != 0 || maximum == 0 {
		return 1
	}
	return 0
}

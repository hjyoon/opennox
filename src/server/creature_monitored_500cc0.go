package server

const (
	creatureMonitoredMonsterClassLow500CC0 = uint8(0x02)
	creatureMonitoredDeadFlag500CC0        = uint32(0x00008000)
	creatureMonitoredSummonedStatus500CC0  = uint8(0x80)
)

// creatureMonitoredHooks500CC0 exposes every observable read and predicate
// call made by GAME.EXE 00500CC0. Object and update-data handles remain at
// native width; only the original low class/status bytes and 32-bit flags are
// interpreted here.
type creatureMonitoredHooks500CC0[O, U comparable] struct {
	loadClassLow  func(O) uint8
	loadFlags     func(O) uint32
	isZombie      func(O) bool
	loadUpdate    func(O) U
	loadStatusLow func(U) uint8
	hasOwner      func(unit, owner O) bool
}

// creatureIsMonitored500CC0 preserves the original short-circuit expression:
// an alive Monster bypasses the zombie predicate, while a dead Monster or any
// non-Monster must be accepted by that predicate. The original has no null
// preflight and therefore begins by reading the unit's class byte.
func creatureIsMonitored500CC0[O, U comparable](owner, unit O, hooks creatureMonitoredHooks500CC0[O, U]) bool {
	if hooks.loadClassLow(unit)&creatureMonitoredMonsterClassLow500CC0 == 0 ||
		hooks.loadFlags(unit)&creatureMonitoredDeadFlag500CC0 != 0 {
		if !hooks.isZombie(unit) {
			return false
		}
	}
	update := hooks.loadUpdate(unit)
	if hooks.loadStatusLow(update)&creatureMonitoredSummonedStatus500CC0 == 0 {
		return false
	}
	return hooks.hasOwner(unit, owner)
}

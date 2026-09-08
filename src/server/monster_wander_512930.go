package server

const (
	monsterWanderMonsterClassLow512930 = uint8(0x02)
	monsterWanderBlockedFlag512930     = uint32(0x00008000)
	monsterWanderReportAction512930    = uint32(32)
	monsterWanderRoamAction512930      = uint32(10)
	monsterWanderReportArg512930       = uint32(10)
)

// monsterWanderHooks512930 exposes every observable load, call, and store
// made by GAME.EXE 00512930. Generic handles keep the object, cached update
// data, and returned action identities at native width.
type monsterWanderHooks512930[O, U any, A comparable] struct {
	loadClassLow     func(O) uint8
	loadFlags        func(O) uint32
	loadUpdate       func(O) U
	clearActionStack func(O)
	pushAction       func(O, uint32) A
	storeArgU32      func(A, int, uint32)
	loadRoamFlagsLow func(U) uint8
	storeArgLow      func(A, int, uint8)
}

// monsterWander512930 replaces an eligible monster's action stack with a
// REPORT marker followed by ROAM. The UpdateData handle is intentionally
// loaded before any stack callback, while Field333 is read through that
// cached handle only after the ROAM argument-zero store. The original has no
// null-object or null-UpdateData guard, so this semantic core adds none.
func monsterWander512930[O, U any, A comparable](unit O, hooks monsterWanderHooks512930[O, U, A]) {
	if hooks.loadClassLow(unit)&monsterWanderMonsterClassLow512930 == 0 {
		return
	}
	if hooks.loadFlags(unit)&monsterWanderBlockedFlag512930 != 0 {
		return
	}
	update := hooks.loadUpdate(unit)
	hooks.clearActionStack(unit)

	report := hooks.pushAction(unit, monsterWanderReportAction512930)
	var nilAction A
	if report != nilAction {
		hooks.storeArgU32(report, 0, monsterWanderReportArg512930)
	}

	roam := hooks.pushAction(unit, monsterWanderRoamAction512930)
	if roam == nilAction {
		return
	}
	hooks.storeArgU32(roam, 0, 0)
	flags := hooks.loadRoamFlagsLow(update)
	hooks.storeArgLow(roam, 2, flags)
}

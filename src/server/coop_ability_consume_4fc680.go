package server

type coopAbilityConsumeHooks4FC680[Unit comparable] struct {
	loadCoopFlag    func() int32
	loadFlag20      func() int32
	loadState       func() int32
	firstPlayerUnit func() Unit
	executeAbility  func(Unit, int32)
	storeState      func(int32)
}

// coopAbilityConsume4FC680 preserves GAME.EXE 004FC680. The queued state is
// tested before player lookup, then reloaded after that callback immediately
// before execution. The complete state dword is cleared only after execution
// returns normally.
func coopAbilityConsume4FC680[Unit comparable](hooks coopAbilityConsumeHooks4FC680[Unit]) {
	if hooks.loadCoopFlag() == 0 {
		return
	}
	if hooks.loadFlag20() == 1 {
		return
	}
	if hooks.loadState() == 0 {
		return
	}
	unit := hooks.firstPlayerUnit()
	var nilUnit Unit
	if unit == nilUnit {
		return
	}
	state := hooks.loadState()
	hooks.executeAbility(unit, state)
	hooks.storeState(0)
}

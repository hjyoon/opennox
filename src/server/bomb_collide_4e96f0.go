package server

type bombCollideHooks4E96F0[O, U, B comparable] struct {
	loadUpdateData   func(O) U
	gameModeCoop     func() int32
	firstPlayerUnit  func() O
	loadFlags        func(O) uint32
	collisionBlock   func(U) B
	scriptCallback   func(B, O, O)
	classLow         func(O) uint8
	unitsOnSameTeam  func(O, O) int32
	storeCollideUnit func(U, O)
	damageClear      func(O, int32)
}

// bombCollide4E96F0 preserves GAME.EXE 004E96F0. The bomb update pointer is
// cached before the Coop gate. A Coop return value of exactly one causes the
// first player unit's live flags to be read; bit 1 suppresses every callback.
// Otherwise script event 21 runs before the other object's live class, owner
// chains, and flags are inspected. On the accepted path, the other object is
// stored through the cached update pointer before 999 damage is applied. The
// registered eight-byte collide record and collision pointer are not read.
func bombCollide4E96F0[O, U, B comparable](
	bomb, other O,
	collision any,
	hooks bombCollideHooks4E96F0[O, U, B],
) {
	_ = collision
	update := hooks.loadUpdateData(bomb)
	if hooks.gameModeCoop() == 1 {
		first := hooks.firstPlayerUnit()
		if hooks.loadFlags(first)&0x2 == 0x2 {
			return
		}
	}

	block := hooks.collisionBlock(update)
	hooks.scriptCallback(block, other, bomb)

	var zero O
	if other == zero {
		return
	}
	if hooks.classLow(other)&0x6 == 0 {
		return
	}
	if hooks.unitsOnSameTeam(bomb, other) != 0 {
		return
	}
	if hooks.loadFlags(other)&0x8000 != 0 {
		return
	}
	hooks.storeCollideUnit(update, other)
	hooks.damageClear(bomb, 999)
}

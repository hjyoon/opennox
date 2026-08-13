package server

type monsterCollideHooks4E83B0[O, U, B, R any] struct {
	updateData     func(O) U
	collisionBlock func(U) B
	scriptCallback func(B, O, O) R
}

// monsterCollide4E83B0 preserves GAME.EXE 004E83B0. The original performs no
// class or nil checks: it obtains the collision callback block from the first
// object's update data, then calls it with the second object as caller and the
// first object as trigger. The callback result is returned without conversion.
func monsterCollide4E83B0[O, U, B, R any](monster, other O, hooks monsterCollideHooks4E83B0[O, U, B, R]) R {
	update := hooks.updateData(monster)
	block := hooks.collisionBlock(update)
	return hooks.scriptCallback(block, other, monster)
}

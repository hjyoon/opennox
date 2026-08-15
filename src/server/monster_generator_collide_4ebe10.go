package server

const monsterGeneratorPlayerClassLow4EBE10 = uint8(0x04)

type monsterGeneratorCollideHooks4EBE10[O, U, B, R any] struct {
	loadTargetClassLow func(O) uint8
	loadSourceUpdate   func(O) U
	collisionBlock     func(U) B
	scriptCallback     func(B, O, O) R
}

// monsterGeneratorCollide4EBE10 preserves GAME.EXE 004EBE10. The original
// inspects only the target until both the nil and low-byte Player gates pass.
// It then loads the source update-data pointer, forms the callback block at
// original offset 72 and calls it with target as caller and source as trigger.
// The registered collision point and callback result are not consumed.
func monsterGeneratorCollide4EBE10[O comparable, U, B, R any](
	source, target O,
	hooks monsterGeneratorCollideHooks4EBE10[O, U, B, R],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.loadTargetClassLow(target)&monsterGeneratorPlayerClassLow4EBE10 == 0 {
		return
	}
	update := hooks.loadSourceUpdate(source)
	block := hooks.collisionBlock(update)
	_ = hooks.scriptCallback(block, target, source)
}

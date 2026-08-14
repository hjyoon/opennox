package server

const fistCollideDamageType4EADF0 = uint32(2)

type fistCollideHooks4EADF0[O comparable, D, F any] struct {
	loadZ            func(O) float32
	loadUpdateData   func(O) D
	loadDamage       func(D) int32
	findParentPlayer func(O) O
	loadTargetDamage func(O) F
	callTargetDamage func(F, O, O, O, int32, uint32) int32
}

// fistCollide4EADF0 preserves GAME.EXE 004EADF0. The x87 height comparison
// accepts zero, negative values, and unordered NaNs; only an ordered value
// strictly greater than positive zero returns early. On the accepted path the
// source update-data pointer is cached before the nil-target guard. Its first
// int32 is read before the live parent lookup, while the target Damage callback
// is loaded after that lookup. The callback result and collision point are
// ignored.
func fistCollide4EADF0[O comparable, D, F, C any](
	source, target O,
	_ C,
	hooks fistCollideHooks4EADF0[O, D, F],
) {
	if hooks.loadZ(source) > 0 {
		return
	}
	data := hooks.loadUpdateData(source)
	var zero O
	if target == zero {
		return
	}
	damage := hooks.loadDamage(data)
	parent := hooks.findParentPlayer(source)
	damageFn := hooks.loadTargetDamage(target)
	_ = hooks.callTargetDamage(
		damageFn,
		target,
		parent,
		source,
		damage,
		fistCollideDamageType4EADF0,
	)
}

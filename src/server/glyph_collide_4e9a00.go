package server

type glyphCollideHooks4E9A00[O comparable] struct {
	allowed func(O, O) int32
	trigger func(O, O)
}

// glyphCollide4E9A00 preserves GAME.EXE 004E9A00. The eligibility helper is
// called before any trap effect, and any nonzero helper result invokes the
// effect with the original source and target arguments. The registered
// collide-data size is zero and the third collision argument is not read.
func glyphCollide4E9A00[O comparable](
	source, target O,
	collision any,
	hooks glyphCollideHooks4E9A00[O],
) {
	_ = collision
	if hooks.allowed(source, target) == 0 {
		return
	}
	hooks.trigger(source, target)
}

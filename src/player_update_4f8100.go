package opennox

// playerUpdateHarpoonHooks4F8100 exposes the observable loads and calls in the
// Harpoon tail of GAME.EXE 004F8100. The update-data pointer itself remains
// cached by the caller, while its target slot is deliberately reloaded around
// callbacks that may mutate it.
type playerUpdateHarpoonHooks4F8100[T comparable] struct {
	loadTarget  func() T
	loadForce   func() float64
	destroyed   func(T) bool
	breakOwner  func()
	attribution func(T)
	applyForce  func(T, float64)
}

// playerUpdateHarpoon4F8100 preserves 004F83B1..004F840C. In particular, the
// force lookup occurs before the destroyed-bit load, and the target is loaded
// again both after that lookup and after attribution. There are intentionally
// no nil checks on either reloaded target: the original immediately
// dereferences or passes those values after the corresponding callback.
func playerUpdateHarpoon4F8100[T comparable](h playerUpdateHarpoonHooks4F8100[T]) {
	target := h.loadTarget()
	var zero T
	if target == zero {
		return
	}

	force := h.loadForce()
	target = h.loadTarget()
	if h.destroyed(target) {
		h.breakOwner()
		return
	}

	h.attribution(target)
	target = h.loadTarget()
	// GAME.EXE spills HarpoonForce through m32real before negating and passing
	// it to ObjectApplyForce, so retain that float32 rounding on native hosts.
	h.applyForce(target, -float64(float32(force)))
}

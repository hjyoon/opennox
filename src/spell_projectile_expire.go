package opennox

import "github.com/opennox/libs/object"

type spellProjectileExpire4E71F0Hooks[T comparable, U any] struct {
	updateData    func(T) U
	search        func(T, float32, *targetSearchArg4E6EA0[T]) T
	level         func(U) int32
	owner         func(U) T
	spell         func(U) int32
	source        func(U) T
	accept        func(int32, T, T, T, T, int32) int32
	delayedDelete func(T)
}

// spellProjectileExpire4E71F0 preserves the observable load and call order of
// GAME.EXE at 004E71F0. The update-data pointer is cached before target search.
// If a target is found, level is read first, owner is read once and reused for
// both owner arguments, then spell and source are read before spell dispatch.
// The spell result is ignored and the projectile is always delayed-delete
// queued after a normal search or dispatch return.
func spellProjectileExpire4E71F0[T comparable, U any](obj T, h spellProjectileExpire4E71F0Hooks[T, U]) {
	ud := h.updateData(obj)
	target := h.search(obj, 50, &targetSearchArg4E6EA0[T]{
		Field0:             15,
		Field4:             1,
		Field8:             0,
		ClassAllow12:       object.MaskUnits,
		ClassDisallow16:    0,
		SubClassAllow20:    object.SubClass(^uint32(0)),
		SubClassDisallow24: 0,
		FlagsAllow28:       object.Flags(^uint32(0)),
		FlagsDisallow32:    object.FlagDead | object.FlagDestroyed,
	})
	var zero T
	if target != zero {
		level := h.level(ud)
		owner := h.owner(ud)
		spellID := h.spell(ud)
		source := h.source(ud)
		_ = h.accept(spellID, source, owner, owner, target, level)
	}
	h.delayedDelete(obj)
}

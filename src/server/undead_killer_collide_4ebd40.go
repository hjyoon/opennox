package server

const (
	undeadKillerMonsterClassLow4EBD40   = uint8(0x02)
	undeadKillerUndeadSubclassLow4EBD40 = uint8(0x40)
	undeadKillerDamageType4EBD40        = uint32(6)
)

type undeadKillerCollideHooks4EBD40[O, D, S, F any] struct {
	loadClassLow     func(O) uint8
	loadSubclassLow  func(O) uint8
	loadCollideData  func(O) D
	loadHP           func(O) uint16
	loadSpell        func(D) S
	loadRemaining    func(S) int32
	findParentPlayer func(O) O
	loadTargetDamage func(O) F
	callTargetDamage func(F, O, O, O, int32, uint32) int32
	delayedDelete    func(O)
	storeRemaining   func(S, int32)
}

// undeadKillerCollide4EBD40 preserves GAME.EXE 004EBD40. A nil target
// deletes the killer only when the collision pointer is also nil. Eligible
// undead monsters consume the kill-point budget stored in the cached duration
// spell. The partial-consumption branch stores from the entry budget, while
// the full-consumption branch reloads the live budget after damage and delete.
func undeadKillerCollide4EBD40[O, C comparable, D, S, F any](
	source, target O,
	collision C,
	hooks undeadKillerCollideHooks4EBD40[O, D, S, F],
) {
	var zeroObject O
	if target == zeroObject {
		var zeroCollision C
		if collision == zeroCollision {
			hooks.delayedDelete(source)
		}
		return
	}
	if hooks.loadClassLow(target)&undeadKillerMonsterClassLow4EBD40 == 0 {
		return
	}
	if hooks.loadSubclassLow(target)&undeadKillerUndeadSubclassLow4EBD40 == 0 {
		return
	}

	data := hooks.loadCollideData(source)
	hp := int32(hooks.loadHP(target))
	spell := hooks.loadSpell(data)
	remaining := hooks.loadRemaining(spell)
	if remaining > hp {
		parent := hooks.findParentPlayer(source)
		damageFn := hooks.loadTargetDamage(target)
		_ = hooks.callTargetDamage(
			damageFn,
			target,
			parent,
			source,
			hp,
			undeadKillerDamageType4EBD40,
		)
		hooks.storeRemaining(spell, remaining-hp)
		return
	}
	if remaining != 0 {
		parent := hooks.findParentPlayer(source)
		damageFn := hooks.loadTargetDamage(target)
		_ = hooks.callTargetDamage(
			damageFn,
			target,
			parent,
			source,
			remaining,
			undeadKillerDamageType4EBD40,
		)
		hooks.delayedDelete(source)
		liveRemaining := hooks.loadRemaining(spell)
		hooks.storeRemaining(spell, liveRemaining-remaining)
		return
	}
	hooks.delayedDelete(source)
}

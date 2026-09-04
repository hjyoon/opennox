package opennox

const spellCancellationPlayerClass4FCEB0 = uint32(0x04)

type spellCancellationTraversalHooks4FCEB0[Spell, Target comparable] struct {
	firstSpell      func() Spell
	loadNext        func(Spell) Spell
	loadTarget      func(Spell) Target
	loadTargetClass func(Target) uint32
	cancelSpell     func(Spell)
}

// spellCancellationTraversal4FCEB0 preserves GAME.EXE 004FCEB0's exact
// traversal and short-circuit order while keeping pointer-bearing values at
// their native width. Each next link is cached before any mode or target
// observation, so cancellation cannot redirect the traversal by changing the
// current record. Only mode 1 preserves spells aimed at Player-class targets.
func spellCancellationTraversal4FCEB0[Spell, Target comparable](
	mode int32,
	hooks spellCancellationTraversalHooks4FCEB0[Spell, Target],
) int32 {
	var nilSpell Spell
	var nilTarget Target
	for current := hooks.firstSpell(); current != nilSpell; {
		next := hooks.loadNext(current)
		if mode != 1 {
			hooks.cancelSpell(current)
		} else {
			target := hooks.loadTarget(current)
			if target == nilTarget || hooks.loadTargetClass(target)&spellCancellationPlayerClass4FCEB0 == 0 {
				hooks.cancelSpell(current)
			}
		}
		current = next
	}
	return 0
}

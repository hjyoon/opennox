package server

const awardSpellCollideGrantMode4EAD20 = int32(1)

type awardSpellCollideHooks4EAD20[O, D any] struct {
	loadCollideData func(O) D
	loadSpell       func(D) uint32
	grantSpell      func(O, uint32, int32, int32, int32) int32
}

// awardSpellCollide4EAD20 preserves GAME.EXE 004EAD20. A nil target returns
// zero before source or collision is observed. Otherwise it loads the source
// collide-data pointer and its full 32-bit spell ID, then returns the complete
// result of grant(target, spell, 1, 0, 0). Collision is never observed.
func awardSpellCollide4EAD20[O comparable, D, C any](
	source, target O,
	_ C,
	hooks awardSpellCollideHooks4EAD20[O, D],
) int32 {
	var zero O
	if target == zero {
		return 0
	}
	data := hooks.loadCollideData(source)
	spell := hooks.loadSpell(data)
	return hooks.grantSpell(target, spell, awardSpellCollideGrantMode4EAD20, 0, 0)
}

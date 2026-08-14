package server

const (
	bearTrapClosedType4EB890      = "ClosedBearTrap"
	bearTrapHeldEnchant4EB890     = uint32(5)
	bearTrapAnchoredEnchant4EB890 = uint32(14)
	bearTrapEnchantDuration4EB890 = uint32(90)
	bearTrapEnchantPower4EB890    = uint32(5)
	bearTrapTriggeredSound4EB890  = uint32(846)
)

type bearTrapCollideHooks4EB890[O comparable] struct {
	allowed       func(O, O) int32
	newObject     func(string) O
	loadPosY      func(O) float32
	loadPosX      func(O) float32
	loadOwner     func(O) O
	createAt      func(O, O, float32, float32, uint32)
	delayedDelete func(O)
	applyEnchant  func(O, uint32, uint32, uint32)
	audio         func(uint32, O, int32, uint32)
}

// bearTrapCollide4EB890 preserves GAME.EXE 004EB890. A nil target returns
// before the source is used. The shared glyph gate and ClosedBearTrap
// allocation must both succeed before any source field is read. On success,
// source Y is cached before X, owner is loaded last, and CreateAt receives the
// extra zero slot present at the original call site. The source is then
// deleted, the target is held and anchored in that order, and the trigger
// sound is emitted on the source. Collision is part of the registered
// three-pointer ABI but is never read.
func bearTrapCollide4EB890[O comparable](
	source, target O,
	_ any,
	hooks bearTrapCollideHooks4EB890[O],
) {
	var zero O
	if target == zero {
		return
	}
	if hooks.allowed(source, target) == 0 {
		return
	}
	closed := hooks.newObject(bearTrapClosedType4EB890)
	if closed == zero {
		return
	}

	y := hooks.loadPosY(source)
	x := hooks.loadPosX(source)
	owner := hooks.loadOwner(source)
	hooks.createAt(closed, owner, x, y, 0)
	hooks.delayedDelete(source)
	hooks.applyEnchant(target, bearTrapHeldEnchant4EB890, bearTrapEnchantDuration4EB890, bearTrapEnchantPower4EB890)
	hooks.applyEnchant(target, bearTrapAnchoredEnchant4EB890, bearTrapEnchantDuration4EB890, bearTrapEnchantPower4EB890)
	hooks.audio(bearTrapTriggeredSound4EB890, source, 0, 0)
}

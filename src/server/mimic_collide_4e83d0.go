package server

const (
	mimicCollideUnitClassMask4E83D0     = uint8(0x06)
	mimicCollideDeadFlag4E83D0          = uint32(0x00008000)
	mimicCollideFightAction4E83D0       = uint32(15)
	mimicCollideUnderAttackAction4E83D0 = uint32(43)
)

type mimicCollideHooks4E83D0[O, A comparable, P, R any] struct {
	flags           func(O) uint32
	classLow        func(O) uint8
	isEnemy         func(O, O) int32
	actionScheduled func(O, uint32) int32
	pushAction      func(O, uint32) A
	frame           func() uint32
	storeActionArg  func(A, int, uint32)
	posXBits        func(O) uint32
	posYBits        func(O) uint32
	monsterCollide  func(O, O, P) R
}

// mimicCollide4E83D0 preserves GAME.EXE 004E83D0. The optional AI work uses
// exact-zero tests and live reads in original order. Regardless of every
// branch, the three original collision arguments are forwarded to 004E83B0
// and its result is returned unchanged.
func mimicCollide4E83D0[O, A comparable, P, R any](
	mimic, other O,
	collision P,
	hooks mimicCollideHooks4E83D0[O, A, P, R],
) R {
	var zeroObject O
	var zeroAction A
	if other != zeroObject {
		flags := hooks.flags(other)
		if flags&mimicCollideDeadFlag4E83D0 == 0 {
			classLow := hooks.classLow(other)
			if classLow&mimicCollideUnitClassMask4E83D0 != 0 {
				if hooks.isEnemy(mimic, other) != 0 &&
					hooks.actionScheduled(mimic, mimicCollideFightAction4E83D0) == 0 {
					underAttack := hooks.pushAction(mimic, mimicCollideUnderAttackAction4E83D0)
					if underAttack != zeroAction {
						frame := hooks.frame()
						hooks.storeActionArg(underAttack, 0, frame)
					}

					fight := hooks.pushAction(mimic, mimicCollideFightAction4E83D0)
					if fight != zeroAction {
						x := hooks.posXBits(other)
						hooks.storeActionArg(fight, 0, x)
						y := hooks.posYBits(other)
						hooks.storeActionArg(fight, 1, y)
						frame := hooks.frame()
						hooks.storeActionArg(fight, 2, frame)
					}
				}
			}
		}
	}
	return hooks.monsterCollide(mimic, other, collision)
}

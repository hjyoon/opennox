package server

const (
	monsterArrowCoopFlag4EB800         = uint32(0x800)
	monsterArrowTargetReject4EB800     = uint32(0x8000)
	monsterArrowTargetDamageType4EB800 = uint32(3)
	monsterArrowMapDamageType4EB800    = uint32(11)
)

type monsterArrowCollideHooks4EB800[O, D, P any] struct {
	loadCollideData func(O) D
	gameFlag        func(uint32) bool
	loadCoopDamage  func(D) int32
	loadOtherDamage func(D) int32
	loadTargetFlags func(O) uint32
	findParent      func(O) O
	targetDamage    func(O, O, O, int32, uint32) int32
	tracePoint      func() P
	loadTraceY      func(P) int32
	loadTraceX      func(P) int32
	damageMap       func(int32, int32, int32, uint32, O)
	delayedDelete   func(O)
}

// monsterArrowCollide4EB800 preserves GAME.EXE 004EB800. Collide-data and
// the selected signed damage are cached before the target branch. A target
// carrying flag 0x8000 is ignored without deleting the source; every other
// target is damaged with type 3 and the Damage result is discarded before
// deleting the source. The wall path reads trace Y before X, damages the map
// with type 11 when a trace point exists, and deletes even without one.
// Collision belongs to the registered three-pointer ABI but is never read.
func monsterArrowCollide4EB800[O comparable, D any, P comparable](
	source, target O,
	_ any,
	hooks monsterArrowCollideHooks4EB800[O, D, P],
) {
	data := hooks.loadCollideData(source)
	var damage int32
	if hooks.gameFlag(monsterArrowCoopFlag4EB800) {
		damage = hooks.loadCoopDamage(data)
	} else {
		damage = hooks.loadOtherDamage(data)
	}

	var zeroObject O
	if target != zeroObject {
		if hooks.loadTargetFlags(target)&monsterArrowTargetReject4EB800 != 0 {
			return
		}
		parent := hooks.findParent(source)
		hooks.targetDamage(target, parent, source, damage, monsterArrowTargetDamageType4EB800)
		hooks.delayedDelete(source)
		return
	}

	point := hooks.tracePoint()
	var zeroPoint P
	if point != zeroPoint {
		y := hooks.loadTraceY(point)
		x := hooks.loadTraceX(point)
		hooks.damageMap(x, y, damage, monsterArrowMapDamageType4EB800, source)
	}
	hooks.delayedDelete(source)
}

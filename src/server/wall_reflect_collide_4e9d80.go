package server

import "math"

const (
	wallReflectQuestFlag4E9D80       = uint32(0x1000)
	wallReflectDamageType4E9D80      = uint32(11)
	wallReflectGridInverseBits4E9D80 = uint32(0x3d321643)
	yellowStarSuppressFXFlag4E9E50   = uint32(0x4)
	yellowStarPointFX4E9E50          = uint32(136)
)

type wallReflectCollideHooks4E9D80[O, D, F comparable, C any] struct {
	loadCollideData   func(O) C
	sameTeam          func(O, O) int32
	gameFlagsCheck    func(uint32) int32
	loadCollide       func(O) F
	yellowStarCollide F
	loadDamage        func(C) int32
	findParent        func(O) O
	targetDamage      func(O, O, O, int32, uint32) int32
	delayedDelete     func(O)
	wallReflect       func(D, O)
	loadNewPosY       func(O) float32
	loadNewPosX       func(O) float32
	floatToInt        func(float32) int32
	damageMap         func(int32, int32, int32, uint32, O)
}

// wallReflectCollide4E9D80 preserves GAME.EXE 004E9D80. The collide-data
// pointer is cached at entry, before either collision branch. Target damage
// is tripled only in Quest mode when the source's live Collide callback is the
// exact YellowStarShot callback. The wall path reflects first, then observes
// live Y, cached-data damage, and live X in the original order.
func wallReflectCollide4E9D80[O, D, F comparable, C any](
	source, target O,
	collision D,
	hooks wallReflectCollideHooks4E9D80[O, D, F, C],
) {
	data := hooks.loadCollideData(source)

	var zeroObject O
	if target != zeroObject {
		if hooks.sameTeam(source, target) != 0 {
			return
		}

		var damage int32
		if hooks.gameFlagsCheck(wallReflectQuestFlag4E9D80) != 0 &&
			hooks.loadCollide(source) == hooks.yellowStarCollide {
			damage = hooks.loadDamage(data) * 3
		} else {
			damage = hooks.loadDamage(data)
		}
		parent := hooks.findParent(source)
		if hooks.targetDamage(target, parent, source, damage, wallReflectDamageType4E9D80) != 0 {
			hooks.delayedDelete(source)
		}
		return
	}

	var zeroCollision D
	if collision == zeroCollision {
		return
	}
	hooks.wallReflect(collision, source)
	y := hooks.loadNewPosY(source)
	damage := hooks.loadDamage(data)
	gridInverse := math.Float32frombits(wallReflectGridInverseBits4E9D80)
	gridY := hooks.floatToInt(y * gridInverse)
	x := hooks.loadNewPosX(source)
	gridX := hooks.floatToInt(x * gridInverse)
	hooks.damageMap(gridX, gridY, damage, wallReflectDamageType4E9D80, source)
}

type yellowStarShotCollideHooks4E9E50[O, D comparable] struct {
	gameFlagsCheck func(uint32) int32
	pointFX        func(uint32, O)
	wallCollide    func(O, O, D)
}

// yellowStarShotCollide4E9E50 preserves GAME.EXE 004E9E50. A nil source
// skips the flag and FX reads but is still forwarded to 004E9D80, where the
// original entry dereference faults. Every non-nil path forwards all three
// callback arguments unchanged.
func yellowStarShotCollide4E9E50[O, D comparable](
	source, target O,
	collision D,
	hooks yellowStarShotCollideHooks4E9E50[O, D],
) {
	var zeroObject O
	if source != zeroObject && hooks.gameFlagsCheck(yellowStarSuppressFXFlag4E9E50) == 0 {
		hooks.pointFX(yellowStarPointFX4E9E50, source)
	}
	hooks.wallCollide(source, target, collision)
}

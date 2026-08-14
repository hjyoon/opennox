package server

import "math"

const (
	boomCollideDamageBalance4E9770       = "MagicMissileDamage"
	boomCollideSplashBalance4E9770       = "MagicMissileSplashDamage"
	boomCollideRangeBalance4E9770        = "MagicMissileRange"
	boomCollidePushRangeBalance4E9770    = "MagicMissilePushRange"
	boomCollideForceBalance4E9770        = "MagicMissileForce"
	boomCollideQuestFlag4E9770           = uint32(0x1000)
	boomCollidePlayerClassLow4E9770      = uint8(0x4)
	boomCollideInversionEnchant4E9770    = uint32(27)
	boomCollidePointFX4E9770             = uint32(134)
	boomCollideReflectAudio4E9770        = uint32(122)
	boomCollideDetonateAudio4E9770       = uint32(84)
	boomCollideExplosionDamageType4E9770 = uint32(7)
	boomCollideScorchType4E9770          = int32(0)
	boomCollideInnerRadius4E9770         = float32(5)
	boomCollideVelocityScaleBits4E9770   = uint32(0x3f000000)
	boomDirectionHalfBits509ED0          = uint32(0x3f000000)
	boomDirectionScaleBits509ED0         = uint32(0x4222f983)
	boomDirectionTauBits509ED0           = uint32(0x40c90fdb)
)

type boomCollideHooks4E9770[O, C, P comparable] struct {
	loadBalanceReady  func() uint32
	gameDataFloat     func(string) float64
	floatToInt        func(float32) int32
	storeDirectDamage func(int32)
	storeSplashDamage func(int32)
	storeRange        func(float32)
	storePushRange    func(float32)
	storeForce        func(float32)
	storeBalanceReady func(uint32)
	gameFlagsCheck    func(uint32) int32
	findParent        func(O) O
	classLow          func(O) uint8
	isEnemy           func(O, O) int32
	pointFX           func(uint32, O)
	inversion         func(O, O) int32
	changeOwner       func(O, O)
	hasEnchant        func(O, uint32) int32
	loadDirection     func(O) int16
	checkDirection    func(O, int16, O) int32
	audio             func(uint32, O, int32, uint32)
	loadDirectDamage  func() int32
	targetDamage      func(O, O, O, int32, uint32) int32
	scorch            func(O, int32)
	wallReflect       func(C, O)
	loadVelocityX     func(O) float32
	loadVelocityY     func(O) float32
	vectorDirection   func(float32, float32) int32
	storeDirection2   func(O, uint16)
	storeVelocityX    func(O, float32)
	storeVelocityY    func(O, float32)
	traceHitPoint     func() P
	loadPointY        func(P) int32
	loadPointX        func(P) int32
	damageMap         func(int32, int32, int32, uint32, O)
	loadSplashDamage  func() int32
	loadRange         func() float32
	mapDamageUnits    func(O, float32, float32, int32, uint32, O, O)
	loadForce         func() float32
	loadPushRange     func() float32
	mapPushUnits      func(O, float32, float32, float32, O, int32, int32)
	delayedDelete     func(O)
}

// boomCollide4E9770 preserves GAME.EXE 004E9770. Besides branch results, the
// hook boundary records the original load/call/store order: balance values are
// initialized once in five ordered reads, mutable globals are reloaded at each
// later use, and velocity/trace fields are read around callbacks at the same
// points as the original x86 instructions. BoomCollide's registered data size
// is zero, so the callback receives only source, target and collision inputs.
func boomCollide4E9770[O, C, P comparable](
	source, target O,
	collision C,
	hooks boomCollideHooks4E9770[O, C, P],
) {
	if hooks.loadBalanceReady() == 0 {
		damage := hooks.floatToInt(float32(hooks.gameDataFloat(boomCollideDamageBalance4E9770)))
		hooks.storeDirectDamage(damage)
		splash := hooks.floatToInt(float32(hooks.gameDataFloat(boomCollideSplashBalance4E9770)))
		hooks.storeSplashDamage(splash)
		hooks.storeRange(float32(hooks.gameDataFloat(boomCollideRangeBalance4E9770)))
		hooks.storePushRange(float32(hooks.gameDataFloat(boomCollidePushRangeBalance4E9770)))
		hooks.storeForce(float32(hooks.gameDataFloat(boomCollideForceBalance4E9770)))
		hooks.storeBalanceReady(1)
	}

	var zeroObject O
	if hooks.gameFlagsCheck(boomCollideQuestFlag4E9770) != 0 {
		parent := hooks.findParent(source)
		if parent != zeroObject &&
			target != zeroObject &&
			hooks.classLow(parent)&boomCollidePlayerClassLow4E9770 != 0 &&
			hooks.classLow(target)&boomCollidePlayerClassLow4E9770 != 0 &&
			hooks.isEnemy(parent, target) == 0 {
			return
		}
	}

	if source != zeroObject {
		hooks.pointFX(boomCollidePointFX4E9770, source)
	}
	if target != zeroObject {
		if hooks.classLow(target)&boomCollidePlayerClassLow4E9770 != 0 {
			if hooks.inversion(target, source) != 0 {
				hooks.changeOwner(source, target)
				return
			}
			if hooks.hasEnchant(target, boomCollideInversionEnchant4E9770) != 0 {
				direction := hooks.loadDirection(target)
				if hooks.checkDirection(target, direction, source)&1 != 0 {
					hooks.changeOwner(source, target)
					hooks.audio(boomCollideReflectAudio4E9770, target, 0, 0)
					return
				}
			}
		}

		damage := hooks.loadDirectDamage()
		parent := hooks.findParent(source)
		_ = hooks.targetDamage(
			target,
			parent,
			source,
			damage,
			boomCollideExplosionDamageType4E9770,
		)
		hooks.scorch(target, boomCollideScorchType4E9770)
	} else {
		var zeroCollision C
		if collision != zeroCollision {
			hooks.wallReflect(collision, source)

			velocityX := hooks.loadVelocityX(source)
			velocityY := hooks.loadVelocityY(source)
			direction := hooks.vectorDirection(velocityX, velocityY)

			scale := math.Float32frombits(boomCollideVelocityScaleBits4E9770)
			halfX := hooks.loadVelocityX(source) * scale
			hooks.storeDirection2(source, uint16(direction))
			hooks.storeVelocityX(source, halfX)
			halfY := hooks.loadVelocityY(source) * scale
			hooks.storeVelocityY(source, halfY)

			point := hooks.traceHitPoint()
			var zeroPoint P
			if point != zeroPoint {
				damage := hooks.loadDirectDamage()
				y := hooks.loadPointY(point)
				x := hooks.loadPointX(point)
				hooks.damageMap(x, y, damage, boomCollideExplosionDamageType4E9770, source)
			}
			return
		}
	}

	splash := hooks.loadSplashDamage()
	radius := hooks.loadRange()
	hooks.mapDamageUnits(
		source,
		radius,
		boomCollideInnerRadius4E9770,
		splash,
		boomCollideExplosionDamageType4E9770,
		source,
		zeroObject,
	)
	force := hooks.loadForce()
	pushRange := hooks.loadPushRange()
	hooks.mapPushUnits(source, pushRange, pushRange, force, source, 0, 0)
	hooks.audio(boomCollideDetonateAudio4E9770, source, 0, 0)
	hooks.delayedDelete(source)
}

// directionFromVector509ED0 restores GAME.EXE 00509ED0. The three constants
// are loaded from their exact binary32 encodings. The expression is rounded
// once to binary32 before the default x87 FISTP round-to-nearest-even step;
// NaN and out-of-range values use x87's integer-indefinite result. Go cannot
// expose x87's 80-bit intermediate format portably, so float64 is the stable
// cross-architecture approximation for the pre-spill expression.
func directionFromVector509ED0(x, y float32) int32 {
	tau := float64(math.Float32frombits(boomDirectionTauBits509ED0))
	scale := float64(math.Float32frombits(boomDirectionScaleBits509ED0))
	half := float64(math.Float32frombits(boomDirectionHalfBits509ED0))
	value := float32((math.Atan2(float64(y), float64(x))+tau)*scale + half)
	result := playerCollideRound4E8460(value)
	if result < 0 {
		adjust := ((uint32(255) - uint32(result)) >> 8) << 8
		result = int32(uint32(result) + adjust)
	}
	if result >= 256 {
		result = int32(uint32(result) - 256*(uint32(result)>>8))
	}
	return result
}

package server

import "math"

const (
	wallReflectSparkDamageType4EA200      = uint32(11)
	wallReflectSparkGridInverseBits4EA200 = uint32(0x3d321643)
)

type wallReflectSparkCollideHooks4EA200[O, C comparable, D any] struct {
	loadCollideData func(O) D
	loadDamage      func(D) int32
	findParent      func(O) O
	targetDamage    func(O, O, O, int32, uint32) int32
	delayedDelete   func(O)
	loadCollisionY  func(C) float32
	loadCollisionX  func(C) float32
	loadVelocityX   func(O) float32
	loadVelocityY   func(O) float32
	storeVelocityX  func(O, float32)
	storeVelocityY  func(O, float32)
	loadNewPosY     func(O) float32
	loadNewPosX     func(O) float32
	floatToInt      func(float32) int32
	damageMap       func(int32, int32, int32, uint32, O)
}

// wallReflectSparkCollide4EA200 preserves GAME.EXE 004EA200. The eight-byte
// collide-data pointer is cached before either collision branch. Wall hits
// evaluate the unspilled x87 collision-normal product and take the negated
// swap only for an ordered strict-positive result; negative, zero, and
// unordered results use the plain swap. Map damage observes Y, cached-data
// damage, the first conversion, and then live X in the original order.
func wallReflectSparkCollide4EA200[O, C comparable, D any](
	source, target O,
	collision C,
	hooks wallReflectSparkCollideHooks4EA200[O, C, D],
) {
	data := hooks.loadCollideData(source)

	var zeroObject O
	if target != zeroObject {
		damage := hooks.loadDamage(data)
		parent := hooks.findParent(source)
		if hooks.targetDamage(target, parent, source, damage, wallReflectSparkDamageType4EA200) != 0 {
			hooks.delayedDelete(source)
		}
		return
	}

	var zeroCollision C
	if collision == zeroCollision {
		return
	}

	collisionY := hooks.loadCollisionY(collision)
	collisionX := hooks.loadCollisionX(collision)
	positive := float64(collisionY)*float64(collisionX) > 0
	velocityX := hooks.loadVelocityX(source)
	velocityY := hooks.loadVelocityY(source)
	if positive {
		hooks.storeVelocityX(source, -velocityY)
		hooks.storeVelocityY(source, -velocityX)
	} else {
		hooks.storeVelocityX(source, velocityY)
		hooks.storeVelocityY(source, velocityX)
	}

	y := hooks.loadNewPosY(source)
	damage := hooks.loadDamage(data)
	gridInverse := math.Float32frombits(wallReflectSparkGridInverseBits4EA200)
	gridY := hooks.floatToInt(y * gridInverse)
	x := hooks.loadNewPosX(source)
	gridX := hooks.floatToInt(x * gridInverse)
	hooks.damageMap(gridX, gridY, damage, wallReflectSparkDamageType4EA200, source)
}

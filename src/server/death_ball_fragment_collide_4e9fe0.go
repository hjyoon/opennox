package server

import "math"

const (
	deathBallFragmentDamage4E9FE0          = int32(20)
	deathBallFragmentDamageType4E9FE0      = uint32(2)
	deathBallFragmentWallAudio4E9FE0       = uint32(37)
	deathBallFragmentGridInverseBits4E9FE0 = uint32(0x3d321643)
)

type deathBallFragmentCollideHooks4E9FE0[O, C comparable] struct {
	findParent    func(O) O
	targetDamage  func(O, O, O, int32, uint32) int32
	wallReflect   func(C, O)
	audio         func(uint32, O)
	loadNewPosY   func(O) float32
	loadNewPosX   func(O) float32
	floatToInt    func(float32) int32
	damageMap     func(int32, int32, int32, uint32, O)
	delayedDelete func(O)
}

// deathBallFragmentCollide4E9FE0 preserves GAME.EXE 004E9FE0. A target
// collision always damages and then deletes the fragment, independently of
// the Damage callback result. A wall collision reflects and damages the map
// without deleting the fragment. A collision with neither target nor wall
// deletes it.
func deathBallFragmentCollide4E9FE0[O, C comparable](
	source, target O,
	collision C,
	hooks deathBallFragmentCollideHooks4E9FE0[O, C],
) {
	var zeroObject O
	if target != zeroObject {
		parent := hooks.findParent(source)
		_ = hooks.targetDamage(
			target,
			parent,
			source,
			deathBallFragmentDamage4E9FE0,
			deathBallFragmentDamageType4E9FE0,
		)
		hooks.delayedDelete(source)
		return
	}

	var zeroCollision C
	if collision != zeroCollision {
		hooks.wallReflect(collision, source)
		hooks.audio(deathBallFragmentWallAudio4E9FE0, source)

		gridInverse := math.Float32frombits(deathBallFragmentGridInverseBits4E9FE0)
		y := hooks.loadNewPosY(source)
		gridY := hooks.floatToInt(y * gridInverse)
		x := hooks.loadNewPosX(source)
		gridX := hooks.floatToInt(x * gridInverse)
		hooks.damageMap(
			gridX,
			gridY,
			deathBallFragmentDamage4E9FE0,
			deathBallFragmentDamageType4E9FE0,
			source,
		)
		return
	}

	hooks.delayedDelete(source)
}

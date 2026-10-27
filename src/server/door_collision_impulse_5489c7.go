package server

import (
	"github.com/opennox/libs/types"
	"math"
)

func collisionSqrt53_5488B0(value float64) float64 {
	if value < 0 {
		return math.Float64frombits(0xfff8000000000000)
	}
	return monsterMoveForceSqrtChop53_50D50C(value)
}

// Both original door roots use this same wide impulse sequence. Velocity and
// penetration are binary32 stack values; the damping/force stays PC53/ToZero
// until the component stores. The coefficient belongs to live legacy C state.
func applyDoorCollisionImpulse5489C7(obj *Object, normal types.Pointf, penetration float32, spring func() float32) {
	mul, add := objectForceMul53_52DF80, objectForceAdd53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	y := -mul(float64(normal.Y), memory(obj.VelVec.Y))
	x := -mul(float64(normal.X), memory(obj.VelVec.X))
	velocity := spill(add(y, x))
	damping := collisionSqrt53_5488B0(mul(mul(float64(obj.Mass), memory(spring())), 4))
	force := mul(mul(damping, float64(velocity)), .25)
	force = add(force, mul(float64(spring()), memory(penetration)))
	impulseY := spill(mul(force, memory(normal.Y)))
	impulseX := spill(mul(force, memory(normal.X)))
	obj.Sub548600(types.Ptf(impulseX, impulseY))
}

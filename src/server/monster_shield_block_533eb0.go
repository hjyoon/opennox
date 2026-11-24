package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// ShieldBlockApproach533EB0 follows GAME.EXE 00533F00..00533FDC. The
// normalized X and distance are binary32 spills. Speed, normalized Y and
// the final alignment sum remain at the startup x87 precision.
func ShieldBlockApproach533EB0(defender, missile, velocity, facing types.Pointf) (distance float32, incoming bool) {
	vx, vy := float64(velocity.X), float64(velocity.Y)
	fx, fy := float64(facing.X), float64(facing.Y)
	projectedX := objectForceMul53_52DF80(vx, fx)
	projectedY := objectForceMul53_52DF80(vy, fy)
	if objectForceAdd53_52DF80(projectedX, projectedY) >= 0 {
		return 0, false
	}

	dx := float64(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(missile.X), float64(defender.X))))
	dy := float64(experienceSpill32TowardZero(objectForceSub53_52DF80(float64(missile.Y), float64(defender.Y))))
	perpY := objectForceMul53_52DF80(fx, dy)
	perpX := objectForceMul53_52DF80(-fy, dx)
	if math.Abs(objectForceAdd53_52DF80(perpX, perpY)) >= 20 {
		return 0, false
	}

	vx2, vy2 := objectForceMul53_52DF80(vx, vx), objectForceMul53_52DF80(vy, vy)
	speed := monsterMoveForceSqrtChop53_50D50C(objectForceAdd53_52DF80(vx2, vy2))
	normalX := float64(experienceSpill32TowardZero(objectForceDiv53_52DF80(vx, speed)))
	normalY := objectForceDiv53_52DF80(vy, speed)
	dx2, dy2 := objectForceMul53_52DF80(dx, dx), objectForceMul53_52DF80(dy, dy)
	distance = experienceSpill32TowardZero(monsterMoveForceSqrtChop53_50D50C(objectForceAdd53_52DF80(dy2, dx2)))

	towardX := -objectForceDiv53_52DF80(dx, float64(distance))
	towardY := -objectForceDiv53_52DF80(dy, float64(distance))
	aimX := objectForceMul53_52DF80(normalX, towardX)
	aimY := objectForceMul53_52DF80(normalY, towardY)
	return distance, objectForceAdd53_52DF80(aimY, aimX) > float64(float32(0.7))
}

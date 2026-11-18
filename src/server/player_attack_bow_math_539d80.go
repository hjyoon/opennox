package server

import "github.com/opennox/libs/types"

// PlayerAttackBowSpawn539D80 preserves the retained radius and the original
// PC53/toward-zero position spills in GAME.EXE.
func PlayerAttackBowSpawn539D80(from types.Pointf, radius float32, direction types.Pointf) types.Pointf {
	distance := pointOnLineAddChop53_57C8A0(float64(radius), 4)
	x := pointOnLineMulChop53_57C8A0(distance, float64(direction.X))
	y := pointOnLineMulChop53_57C8A0(distance, float64(direction.Y))
	return types.Ptf(
		pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(x, float64(from.X))),
		pointOnLineSpill57C8A0(pointOnLineAddChop53_57C8A0(y, float64(from.Y))),
	)
}

func PlayerAttackBowVelocity539D80(direction, speed float32) float32 {
	return pointOnLineSpill57C8A0(pointOnLineMulChop53_57C8A0(float64(direction), float64(speed)))
}

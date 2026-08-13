package legacy

import (
	"math"

	"github.com/opennox/opennox/v1/server"
)

const objectDistanceMinimum4E6C00 = float32(0.01)

// objectDistanceShapeExtent_4E6C00 preserves the asymmetric x87 box maximum.
// GAME.EXE keeps width/2 in the x87 register but rounds height/2 through a
// float32 stack slot, and it chooses width only for an ordered, strict greater
// comparison.
func objectDistanceShapeExtent_4E6C00(shape *server.Shape) float64 {
	switch shape.Kind {
	case server.ShapeKindCircle:
		return float64(shape.Circle.R)
	case server.ShapeKindBox:
		width := float64(shape.Box.W) * 0.5
		height := shape.Box.H * 0.5
		if width > float64(height) {
			return width
		}
		return float64(height)
	default:
		return 0
	}
}

// objectDistance_4E6C00 returns the distance between object collision
// surfaces. Position and shape values originate as float32 in GAME.EXE; using
// float64 here preserves their exact differences and products before sqrt.
func objectDistance_4E6C00(a, b *server.Object) float64 {
	dx := float64(a.PosVec.X) - float64(b.PosVec.X)
	dy := float64(a.PosVec.Y) - float64(b.PosVec.Y)
	distance := math.Sqrt(dx*dx + dy*dy)
	distance -= objectDistanceShapeExtent_4E6C00(&a.Shape)
	distance -= objectDistanceShapeExtent_4E6C00(&b.Shape)
	minimum := float64(objectDistanceMinimum4E6C00)
	// GAME.EXE tests x87 C0 only. C0 is set for both less-than and
	// unordered comparisons, so a surviving NaN is clamped as well.
	if math.IsNaN(distance) || distance < minimum {
		return minimum
	}
	return distance
}

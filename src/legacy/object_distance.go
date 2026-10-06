package legacy

import (
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
// surfaces with the same retained x87 precision contract as the server entry.
func objectDistance_4E6C00(a, b *server.Object) float64 {
	// Keep the exported double result and native objects on the same
	// precision-53/ToZero path as MainAI, including both shape subtractions.
	return server.ObjectDistance4E6C00(a, b)
}

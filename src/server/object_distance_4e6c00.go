package server

import "math"

// objectDistanceShapeExtent4E6C00 preserves the asymmetric x87 box maximum:
// width/2 remains unrounded, height/2 is spilled to binary32, and width wins
// only on an ordered, strict comparison.
func objectDistanceShapeExtent4E6C00(shape *Shape) float64 {
	switch shape.Kind {
	case ShapeKindCircle:
		return float64(shape.Circle.R)
	case ShapeKindBox:
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

// ObjectDistance4E6C00 measures collision surfaces, not object centers.
// The original clamps both negative distances and unordered results to 0.01.
func ObjectDistance4E6C00(a, b *Object) float64 {
	dx := float64(a.PosVec.X) - float64(b.PosVec.X)
	dy := float64(a.PosVec.Y) - float64(b.PosVec.Y)
	distance := math.Sqrt(dx*dx + dy*dy)
	distance -= objectDistanceShapeExtent4E6C00(&a.Shape)
	distance -= objectDistanceShapeExtent4E6C00(&b.Shape)
	minimum := float64(float32(0.01))
	if math.IsNaN(distance) || distance < minimum {
		return minimum
	}
	return distance
}

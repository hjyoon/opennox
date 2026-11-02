package server

import "github.com/opennox/libs/types"

// PointInShapeNative547DB0 retains the original NewPos-based raw predicate.
// Circle differences stay at PC53 precision; box sums and unordered gates
// share the independently restored 57B850 arithmetic.
func (obj *Object) PointInShapeNative547DB0(point *types.Pointf) bool {
	switch obj.Shape.Kind {
	case ShapeKindCircle:
		x := objectForceSub53_52DF80(float64(obj.NewPos.X), objectForceMemory32_52DF80(point.X))
		y := objectForceSub53_52DF80(float64(obj.NewPos.Y), objectForceMemory32_52DF80(point.Y))
		y2 := objectForceMul53_52DF80(y, y)
		x2 := objectForceMul53_52DF80(x, x)
		distance2 := objectForceAdd53_52DF80(y2, x2)
		return !(distance2 > objectForceMemory32_52DF80(obj.Shape.Circle.R2))
	case ShapeKindBox:
		return MapPointInBox57B850(&obj.NewPos, &obj.Shape, point)
	default:
		return false
	}
}

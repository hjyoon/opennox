package server

import "github.com/opennox/libs/types"

func PlayerCursorHeight54AFB0(obj *Object) float64 {
	return objectForceAdd53_52DF80(float64(obj.ZVal), objectForceMemory32_52DF80(obj.PosVec.Y))
}

// PlayerCursorGeometry54AFB0 tests the projected shape used by the server's
// cursor picker. All native field stores keep the original binary32 spills.
func PlayerCursorGeometry54AFB0(obj *Object, cursor *types.Pointf, radiusToInt func(float32) int32) bool {
	add, sub, mul := objectForceAdd53_52DF80, objectForceSub53_52DF80, objectForceMul53_52DF80
	spill, memory := experienceSpill32TowardZero, objectForceMemory32_52DF80
	y := spill(sub(float64(obj.PosVec.Y), memory(obj.ZVal)))
	switch obj.Shape.Kind {
	case 2:
		radius := obj.Shape.Circle.R
		radiusSquared := radiusToInt(spill(mul(float64(radius), memory(radius))))
		if cursor.Y > y || cursor.Y < y {
			dx, dy := sub(float64(cursor.X), memory(obj.PosVec.X)), sub(float64(cursor.Y), memory(y))
			return !(float64(radiusSquared) <= add(mul(dy, dy), mul(dx, dx)))
		}
		return cursor.X > spill(sub(float64(obj.PosVec.X), memory(radius))) && cursor.X < spill(add(float64(radius), memory(obj.PosVec.X)))
	case 3:
		center := types.Ptf(obj.PosVec.X, y)
		if MapPointInBox57B850(&center, &obj.Shape, cursor) {
			return true
		}
		b := &obj.Shape.Box
		return add(float64(b.LeftBottom2), memory(obj.PosVec.X)) < float64(cursor.X) && cursor.X < obj.PosVec.X &&
			add(float64(y), memory(b.LeftTop2)) < float64(cursor.Y) && add(float64(y), memory(b.LeftTop2)) > float64(cursor.Y) ||
			cursor.X >= obj.PosVec.X && add(float64(b.RightTop), memory(obj.PosVec.X)) > float64(cursor.X) &&
				add(float64(y), memory(b.RightBottom)) < float64(cursor.Y) && add(float64(y), memory(b.RightBottom)) > float64(cursor.Y)
	}
	return false
}

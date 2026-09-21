package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type mapTraceObstaclesHooks50B580 struct {
	eachObject  func(types.Rectf, func(*Object))
	isEnemy     func(*Object, *Object) bool
	pointOnLine func(types.Pointf, types.Pointf, types.Pointf) (types.Pointf, bool)
	lineTrace   func(types.Rectf, types.Rectf) bool
}

// mapTraceObstacles50B580 is the pointer-width-independent model of GAME.EXE
// 0050B580 and its callback at 0050B600. The original rectangle iterator
// keeps visiting objects after the callback clears its global result; later
// callbacks become no-ops but still advance the iterator's visitation state.
func mapTraceObstacles50B580(from *Object, p1, p2 types.Pointf, hooks mapTraceObstaclesHooks50B580) bool {
	line := types.RectFromPointsf(p1, p2)
	searching := true
	hooks.eachObject(line, func(it *Object) {
		if !searching || from == it {
			return
		}
		class := it.Class()
		if class.HasAny(object.MaskUnits) {
			if hooks.isEnemy(from, it) {
				return
			}
		} else if !class.HasAny(object.ClassImmobile | object.ClassObstacle) {
			return
		}
		if it.Flags().HasAny(object.FlagNoCollide|object.FlagAllowOverlap) || it.Class().Has(object.ClassDoor) {
			return
		}

		pos := it.Pos()
		shape := &it.Shape
		switch shape.Kind {
		case ShapeKindCircle:
			projected, ok := hooks.pointOnLine(p1, p2, it.PosVec)
			if !ok {
				return
			}
			// GAME.EXE subtracts the binary32 coordinates and performs both
			// products and the sum in the x87 register stack. Binary64 exactly
			// represents those operations for binary32 inputs and avoids the
			// premature binary32 rounding of the previous implementation.
			dx := float64(projected.X) - float64(pos.X)
			dy := float64(projected.Y) - float64(pos.Y)
			if dy*dy+dx*dx <= float64(shape.Circle.R2) {
				searching = false
			}
		case ShapeKindBox:
			edges := [...]types.Rectf{
				types.RectFromPointsf(
					pos.Add(types.Ptf(shape.Box.LeftTop, shape.Box.LeftBottom)),
					pos.Add(types.Ptf(shape.Box.LeftBottom2, shape.Box.LeftTop2)),
				),
				types.RectFromPointsf(
					pos.Add(types.Ptf(shape.Box.LeftTop, shape.Box.LeftBottom)),
					pos.Add(types.Ptf(shape.Box.RightTop, shape.Box.RightBottom)),
				),
				types.RectFromPointsf(
					pos.Add(types.Ptf(shape.Box.RightBottom2, shape.Box.RightTop2)),
					pos.Add(types.Ptf(shape.Box.RightTop, shape.Box.RightBottom)),
				),
				types.RectFromPointsf(
					pos.Add(types.Ptf(shape.Box.RightBottom2, shape.Box.RightTop2)),
					pos.Add(types.Ptf(shape.Box.LeftBottom2, shape.Box.LeftTop2)),
				),
			}
			for _, edge := range edges {
				if hooks.lineTrace(line, edge) {
					searching = false
					break
				}
			}
		}
	})
	return searching
}

func (s *Server) MapTraceObstacles(from *Object, p1, p2 types.Pointf) bool {
	return mapTraceObstacles50B580(from, p1, p2, mapTraceObstaclesHooks50B580{
		eachObject: func(rect types.Rectf, fn func(*Object)) {
			s.Map.EachObjInRect(rect, func(it *Object) bool {
				fn(it)
				return true
			})
		},
		isEnemy:     s.IsEnemyTo,
		pointOnLine: PointOnTheLine,
		lineTrace:   LineTraceXxx,
	})
}

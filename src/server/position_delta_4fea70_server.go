package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

func positionDeltaNative4FEA70(object *Object, point *types.Pointf) int32 {
	return positionDelta4FEA70(object, point, positionDeltaHooks4FEA70[*Object, *types.Pointf]{
		loadPointX: func(point *types.Pointf) float32 {
			return point.X
		},
		loadObjectX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPointY: func(point *types.Pointf) float32 {
			return point.Y
		},
		loadObjectY: func(object *Object) float32 {
			return object.PosVec.Y
		},
	})
}

// PositionDelta4FEA70 binds GAME.EXE 004FEA70 to native-width Object and
// Pointf pointers. It deliberately adds no nil guard because the original
// routine faults on the first invalid coordinate load.
//
//go:noinline
func (*Server) PositionDelta4FEA70(object *Object, point *types.Pointf) int32 {
	return positionDeltaNative4FEA70(object, point)
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(types.Pointf{})]
	_ = [1]struct{}{}[4-unsafe.Sizeof(types.Pointf{}.X)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(types.Pointf{}.Y)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.PosVec.X)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.PosVec.Y)]
)

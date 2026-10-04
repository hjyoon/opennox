package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func newDirectedVisionMap(t *testing.T) *Server {
	t.Helper()
	s := &Server{}
	s.Map.Init()
	s.Walls.byPos = make([]*Wall, wallsPerBucket*WallGridSize)
	t.Cleanup(s.Map.Free)
	return s
}

func directedVisionUnit(pos types.Pointf) *Object {
	u := &Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagActive, PosVec: pos, NewPos: pos}
	u.Shape.Kind = ShapeKindCircle
	u.Shape.Circle.R, u.Shape.Circle.R2 = 4, 16
	return u
}

func TestMapTraceVisionDirectedIndexedObstacles(t *testing.T) {
	for _, descending := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			for _, kind := range []string{"box-on-ray", "box-off-ray", "box-not-shadow", "circle-on-ray", "circle-off-ray", "door-on-ray", "door-off-ray", "door-open-subclass", "unsupported-shape"} {
				t.Run(fmt.Sprintf("descending-%t/reverse-%t/%s", descending, reverse, kind), func(t *testing.T) {
					s := newDirectedVisionMap(t)
					from, to := types.Ptf(300, 300), types.Ptf(428, 396)
					if descending {
						from.Y, to.Y = to.Y, from.Y
					}
					point := from.Add(to.Sub(from).Mul(0.25))
					want := true
					it := &Object{ObjClass: object.ClassObstacle, ObjFlags: object.FlagActive | object.FlagShadow, PosVec: point, NewPos: point}
					it.Shape.Kind = ShapeKindBox
					it.Shape.Box.W, it.Shape.Box.H = 16, 16
					it.Shape.Box.Calc()
					switch kind {
					case "box-on-ray":
						want = false
					case "box-off-ray":
						it.PosVec.Y = 696 - point.Y
					case "box-not-shadow":
						it.ObjFlags &^= object.FlagShadow
					case "circle-on-ray", "circle-off-ray":
						it.Shape.Kind = ShapeKindCircle
						it.Shape.Circle.R, it.Shape.Circle.R2 = 8, 64
						want = kind == "circle-off-ray"
						if want {
							it.PosVec.Y = 696 - point.Y
						}
					case "door-on-ray", "door-off-ray", "door-open-subclass":
						it.ObjClass = object.ClassDoor
						var dir byte = 8
						if descending {
							dir = 16
						}
						dims := DoorSize(dir)
						it.PosVec = point.Sub(types.Ptf(float32(dims.X)*0.4375, float32(dims.Y)*0.4375))
						data := new([4]uint32)
						data[3] = uint32(dir)
						it.UpdateData = unsafe.Pointer(data)
						it.Shape.Kind = ShapeKindCircle
						it.Shape.Circle.R, it.Shape.Circle.R2 = 32, 1024
						want = kind != "door-on-ray"
						if kind == "door-off-ray" {
							it.PosVec.Y = 696 - point.Y - float32(dims.Y)*0.4375
						}
						if kind == "door-open-subclass" {
							it.ObjSubClass = 4
						}
					case "unsupported-shape":
						it.Shape.Kind = ShapeKindNone
					}
					it.NewPos = it.PosVec
					u, target := directedVisionUnit(from), directedVisionUnit(to)
					if reverse {
						u, target = target, u
					}
					s.Map.AddObjectToIndex(u)
					s.Map.AddObjectToIndex(target)
					s.Map.AddObjectToIndex(it)
					if !s.MapTraceRayAt(u.PosVec, target.PosVec, nil, nil, 9) {
						t.Fatal("empty wall map fixture is blocked")
					}
					if got := s.MapTraceVision(u, target); got != want {
						t.Fatalf("indexed visibility=%t want=%t source=%v target=%v obstacle=%v shape=%v", got, want, u.PosVec, target.PosVec, it.PosVec, it.Shape.Kind)
					}
					if got := s.CanInteract(u, target, 0); got != want {
						t.Fatalf("CanInteract=%t want=%t", got, want)
					}
				})
			}
		}
	}
}

func TestCanInteractPreservesBlindAndInvisibilityGates(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		sourceBuff, targetBuff uint32
		flags                  int
		want                   bool
	}{
		{"ordinary", 0, 0, 0, true}, {"blinded", 1 << ENCHANT_BLINDED, 0, 0, false},
		{"invisible-target", 0, 1 << ENCHANT_INVISIBLE, 0, false}, {"explicit-invisible-admission", 0, 1 << ENCHANT_INVISIBLE, 1, true},
		{"blind-still-excluded", 1 << ENCHANT_BLINDED, 1 << ENCHANT_INVISIBLE, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newDirectedVisionMap(t)
			u, target := directedVisionUnit(types.Ptf(300, 412)), directedVisionUnit(types.Ptf(412, 300))
			u.Buffs, target.Buffs = tc.sourceBuff, tc.targetBuff
			if got := s.CanInteract(u, target, tc.flags); got != tc.want {
				t.Fatalf("CanInteract=%t want=%t", got, tc.want)
			}
		})
	}
}

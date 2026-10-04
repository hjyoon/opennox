package server

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMapTraceObstaclesDirectedIndexedGeometry(t *testing.T) {
	for _, descending := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			for _, kind := range []string{"box-on-ray", "box-off-ray", "circle-on-ray", "circle-off-ray", "no-collide", "allow-overlap", "door", "unsupported-shape"} {
				t.Run(fmt.Sprintf("descending-%t/reverse-%t/%s", descending, reverse, kind), func(t *testing.T) {
					s := newDirectedVisionMap(t)
					from, to := types.Ptf(300, 300), types.Ptf(428, 396)
					if descending {
						from.Y, to.Y = to.Y, from.Y
					}
					point := from.Add(to.Sub(from).Mul(0.25))
					it := &Object{ObjClass: object.ClassObstacle | object.ClassImmobile, ObjFlags: object.FlagActive, PosVec: point, NewPos: point}
					it.Shape.Kind = ShapeKindBox
					it.Shape.Box.W, it.Shape.Box.H = 16, 16
					it.Shape.Box.Calc()
					want := false
					switch kind {
					case "box-off-ray", "circle-off-ray":
						it.PosVec.Y = 696 - point.Y
						want = true
					case "no-collide":
						it.ObjFlags |= object.FlagNoCollide
						want = true
					case "allow-overlap":
						it.ObjFlags |= object.FlagAllowOverlap
						want = true
					case "door":
						it.ObjClass |= object.ClassDoor
						want = true
					case "unsupported-shape":
						it.Shape.Kind = ShapeKindNone
						want = true
					}
					if kind == "circle-on-ray" || kind == "circle-off-ray" {
						it.Shape.Kind = ShapeKindCircle
						it.Shape.Circle.R, it.Shape.Circle.R2 = 8, 64
					}
					it.NewPos = it.PosVec
					s.Map.AddObjectToIndex(it)
					if reverse {
						from, to = to, from
					}
					if got := s.MapTraceObstacles(directedVisionUnit(from), from, to); got != want {
						t.Fatalf("ranged path clear=%t want=%t from=%v to=%v obstacle=%v", got, want, from, to, it.PosVec)
					}
				})
			}
		}
	}
}

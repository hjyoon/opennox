package server

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func circleObstacle50B580(class object.Class, flags object.Flags) *Object {
	obj := &Object{
		ObjClass: class,
		ObjFlags: flags,
		PosVec:   types.Ptf(10, 20),
	}
	obj.Shape.Kind = ShapeKindCircle
	obj.Shape.Circle.R2 = 1
	return obj
}

func traceOneObstacle50B580(t *testing.T, from, candidate *Object, enemy bool) (clear bool, enemyCalls int, gotRect types.Rectf) {
	t.Helper()
	clear = mapTraceObstacles50B580(from, types.Ptf(30, 40), types.Ptf(10, 20), mapTraceObstaclesHooks50B580{
		eachObject: func(rect types.Rectf, fn func(*Object)) {
			gotRect = rect
			fn(candidate)
		},
		isEnemy: func(gotFrom, gotCandidate *Object) bool {
			enemyCalls++
			if gotFrom != from || gotCandidate != candidate {
				t.Fatalf("isEnemy arguments = %p, %p; want %p, %p", gotFrom, gotCandidate, from, candidate)
			}
			return enemy
		},
		pointOnLine: func(_, _, center types.Pointf) (types.Pointf, bool) {
			return center, true
		},
		lineTrace: func(types.Rectf, types.Rectf) bool {
			t.Fatal("circle candidate called box line trace")
			return false
		},
	})
	return clear, enemyCalls, gotRect
}

func TestMapTraceObstacles50B580FiltersExactCandidateClasses(t *testing.T) {
	from := circleObstacle50B580(object.ClassMonster, 0)
	cases := []struct {
		name           string
		candidate      *Object
		enemy          bool
		wantClear      bool
		wantEnemyCalls int
	}{
		{name: "source", candidate: from, wantClear: true},
		{name: "enemy unit", candidate: circleObstacle50B580(object.ClassPlayer, 0), enemy: true, wantClear: true, wantEnemyCalls: 1},
		{name: "friendly unit", candidate: circleObstacle50B580(object.ClassPlayer, 0), wantClear: false, wantEnemyCalls: 1},
		{name: "plain object", candidate: circleObstacle50B580(0, 0), wantClear: true},
		{name: "immobile", candidate: circleObstacle50B580(object.ClassImmobile, 0), wantClear: false},
		{name: "obstacle", candidate: circleObstacle50B580(object.ClassObstacle, 0), wantClear: false},
		{name: "door", candidate: circleObstacle50B580(object.ClassImmobile|object.ClassDoor, 0), wantClear: true},
		{name: "no collide", candidate: circleObstacle50B580(object.ClassImmobile, object.FlagNoCollide), wantClear: true},
		{name: "allow overlap", candidate: circleObstacle50B580(object.ClassImmobile, object.FlagAllowOverlap), wantClear: true},
		{name: "unsupported shape", candidate: &Object{ObjClass: object.ClassImmobile}, wantClear: true},
	}
	wantRect := types.Rectf{Min: types.Ptf(10, 20), Max: types.Ptf(30, 40)}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clear, enemyCalls, gotRect := traceOneObstacle50B580(t, from, tc.candidate, tc.enemy)
			if clear != tc.wantClear || enemyCalls != tc.wantEnemyCalls {
				t.Fatalf("clear = %t, enemy calls = %d; want %t, %d", clear, enemyCalls, tc.wantClear, tc.wantEnemyCalls)
			}
			if gotRect != wantRect {
				t.Fatalf("query rect = %+v, want %+v", gotRect, wantRect)
			}
		})
	}
}

func TestMapTraceObstacles50B580KeepsIteratorRunningAfterHit(t *testing.T) {
	objects := []*Object{
		circleObstacle50B580(object.ClassImmobile, 0),
		circleObstacle50B580(object.ClassImmobile, 0),
	}
	visited := 0
	projections := 0
	clear := mapTraceObstacles50B580(new(Object), types.Pointf{}, types.Ptf(10, 10), mapTraceObstaclesHooks50B580{
		eachObject: func(_ types.Rectf, fn func(*Object)) {
			for _, obj := range objects {
				visited++
				fn(obj)
			}
		},
		isEnemy: func(*Object, *Object) bool { return false },
		pointOnLine: func(_, _, center types.Pointf) (types.Pointf, bool) {
			projections++
			return center, true
		},
		lineTrace: func(types.Rectf, types.Rectf) bool { return false },
	})
	if clear || visited != 2 || projections != 1 {
		t.Fatalf("clear = %t, visited = %d, projections = %d; want false, 2, 1", clear, visited, projections)
	}
}

func TestMapTraceObstacles50B580CircleUsesExtendedPrecision(t *testing.T) {
	obj := circleObstacle50B580(object.ClassImmobile, 0)
	obj.PosVec = types.Pointf{}
	outside := types.Ptf(1, 0.0001)
	dx, dy := outside.X-obj.PosVec.X, outside.Y-obj.PosVec.Y
	if got := dx*dx + dy*dy; got != obj.Shape.Circle.R2 {
		t.Fatalf("binary32 precondition = %g, want premature rounding to %g", got, obj.Shape.Circle.R2)
	}

	projected := outside
	hooks := mapTraceObstaclesHooks50B580{
		eachObject: func(_ types.Rectf, fn func(*Object)) { fn(obj) },
		isEnemy:    func(*Object, *Object) bool { return false },
		pointOnLine: func(types.Pointf, types.Pointf, types.Pointf) (types.Pointf, bool) {
			return projected, true
		},
		lineTrace: func(types.Rectf, types.Rectf) bool { return false },
	}
	if clear := mapTraceObstacles50B580(new(Object), types.Pointf{}, types.Ptf(2, 0), hooks); !clear {
		t.Fatal("point just outside radius was blocked after binary32 rounding")
	}
	projected = types.Ptf(1, 0)
	if clear := mapTraceObstacles50B580(new(Object), types.Pointf{}, types.Ptf(2, 0), hooks); clear {
		t.Fatal("point exactly on radius was not blocked")
	}
}

func TestMapTraceObstacles50B580TestsBoxEdgesInOriginalOrder(t *testing.T) {
	obj := &Object{ObjClass: object.ClassObstacle, PosVec: types.Ptf(100, 200)}
	obj.Shape.Kind = ShapeKindBox
	obj.Shape.Box.LeftTop = 1
	obj.Shape.Box.LeftBottom = 2
	obj.Shape.Box.LeftBottom2 = 3
	obj.Shape.Box.LeftTop2 = 4
	obj.Shape.Box.RightTop = 5
	obj.Shape.Box.RightBottom = 6
	obj.Shape.Box.RightBottom2 = 7
	obj.Shape.Box.RightTop2 = 8

	wantLine := types.Rectf{Min: types.Ptf(1, 2), Max: types.Ptf(9, 8)}
	wantEdges := []types.Rectf{
		{Min: types.Ptf(101, 202), Max: types.Ptf(103, 204)},
		{Min: types.Ptf(101, 202), Max: types.Ptf(105, 206)},
		{Min: types.Ptf(105, 206), Max: types.Ptf(107, 208)},
	}
	var gotEdges []types.Rectf
	clear := mapTraceObstacles50B580(new(Object), types.Ptf(9, 8), types.Ptf(1, 2), mapTraceObstaclesHooks50B580{
		eachObject: func(_ types.Rectf, fn func(*Object)) { fn(obj) },
		isEnemy:    func(*Object, *Object) bool { return false },
		pointOnLine: func(types.Pointf, types.Pointf, types.Pointf) (types.Pointf, bool) {
			t.Fatal("box candidate called circle projection")
			return types.Pointf{}, false
		},
		lineTrace: func(line, edge types.Rectf) bool {
			if line != wantLine {
				t.Fatalf("line = %+v, want %+v", line, wantLine)
			}
			gotEdges = append(gotEdges, edge)
			return len(gotEdges) == 3
		},
	})
	if clear {
		t.Fatal("third box edge hit did not block trace")
	}
	if !reflect.DeepEqual(gotEdges, wantEdges) {
		t.Fatalf("edges = %+v, want %+v", gotEdges, wantEdges)
	}
}

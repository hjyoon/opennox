package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestMapPushUnitsAround52E040FiltersAndAttenuates(t *testing.T) {
	type candidate struct {
		position types.Pointf
		movable  bool
		visible  bool
	}
	origin := types.Ptf(100, 200)
	candidates := []*candidate{
		{position: types.Ptf(105, 200), movable: true, visible: true},
		{position: types.Ptf(130, 200), movable: true, visible: true},
		{position: types.Ptf(151, 200), movable: true, visible: true},
		{position: types.Ptf(105, 200), visible: true},
		{position: types.Ptf(106, 200), movable: true},
	}
	var gotRect types.Rectf
	var pushed []*candidate
	var forces []float64
	if !mapPushUnitsAround52E040(origin, 50, 10, 100, mapPushUnitsAroundHooks52E040[*candidate]{
		eachInRect: func(rect types.Rectf, fn func(*candidate) bool) {
			gotRect = rect
			for _, it := range candidates {
				if !fn(it) {
					return
				}
			}
		},
		isMovable: func(it *candidate) bool { return it.movable },
		position:  func(it *candidate) types.Pointf { return it.position },
		traceRay: func(_ types.Pointf, to types.Pointf) bool {
			for _, it := range candidates {
				if it.position == to && it.movable {
					return it.visible
				}
			}
			return false
		},
		applyForce: func(it *candidate, from types.Pointf, force float64) {
			if from != origin {
				t.Fatalf("force origin = %v, want %v", from, origin)
			}
			pushed = append(pushed, it)
			forces = append(forces, force)
		},
	}) {
		t.Fatal("native radial push was not handled")
	}
	if gotRect != (types.Rectf{Min: types.Ptf(50, 150), Max: types.Ptf(150, 250)}) {
		t.Fatalf("query rect = %v", gotRect)
	}
	if len(pushed) != 2 || pushed[0] != candidates[0] || pushed[1] != candidates[1] {
		t.Fatalf("pushed candidates = %v", pushed)
	}
	if forces[0] != 100 {
		t.Fatalf("inner force = %v, want 100", forces[0])
	}
	want := float64(float32(100) * (1 - (float32(30.1)-10)/(50-10)))
	if math.Abs(forces[1]-want) > 1e-6 {
		t.Fatalf("attenuated force = %.9f, want %.9f", forces[1], want)
	}
}

func TestMapPushUnitsAround52E040UsesLargerRadiusForDistance(t *testing.T) {
	called := false
	if !mapPushUnitsAround52E040(types.Pointf{}, 10, 20, 7, mapPushUnitsAroundHooks52E040[int]{
		eachInRect: func(rect types.Rectf, fn func(int) bool) {
			if rect != (types.Rectf{Min: types.Ptf(-10, -10), Max: types.Ptf(10, 10)}) {
				t.Fatalf("query rect = %v", rect)
			}
			fn(1)
		},
		isMovable: func(int) bool { return true },
		position:  func(int) types.Pointf { return types.Ptf(15, 0) },
		traceRay:  func(types.Pointf, types.Pointf) bool { return true },
		applyForce: func(int, types.Pointf, float64) {
			called = true
		},
	}) {
		t.Fatal("native radial push was not handled")
	}
	if !called {
		t.Fatal("larger inner radius was not honored by the distance gate")
	}
}

func TestMapPushUnitsAround52E040RejectsMissingHooks(t *testing.T) {
	if mapPushUnitsAround52E040(types.Pointf{}, 1, 0, 1, mapPushUnitsAroundHooks52E040[int]{}) {
		t.Fatal("missing dependencies were accepted")
	}
	if new(Server).MapPushUnitsAround52E040(types.Pointf{}, 1, 0, 1, MapPushUnitsAroundRuntime52E040{}) {
		t.Fatal("missing force callback was accepted")
	}
}

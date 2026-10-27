package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestMapPushUnitsAround52E040FiltersAndAttenuates(t *testing.T) {
	type candidate struct {
		position types.Pointf
		force    types.Pointf
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
	if !mapPushUnitsAround52E040(&origin, 50, 10, 100, mapPushUnitsAroundHooks52E040[*candidate]{
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
		mass:      func(*candidate) float32 { return 1 },
		isMissile: func(*candidate) bool { return false },
		loadForce: func(it *candidate) types.Pointf { return it.force },
		wake:      func(*candidate) {},
		storeForce: func(it *candidate, from types.Pointf, force types.Pointf) {
			if from != origin {
				t.Fatalf("force origin = %v, want %v", from, origin)
			}
			pushed = append(pushed, it)
			forces = append(forces, float64(force.X))
			it.force = force
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
	// Complete original historical-filter-attenuation row, mass 1.
	if forces[0] != float64(math.Float32frombits(1148524825)) {
		t.Fatalf("inner force = %v, want original force X", forces[0])
	}
	want := float64(math.Float32frombits(1140321394))
	if forces[1] != want {
		t.Fatalf("attenuated force = %.9f, want %.9f", forces[1], want)
	}
}

func TestMapPushUnitsAround52E040UsesLargerRadiusForDistance(t *testing.T) {
	called := false
	if !mapPushUnitsAround52E040(&types.Pointf{}, 10, 20, 7, mapPushUnitsAroundHooks52E040[int]{
		eachInRect: func(rect types.Rectf, fn func(int) bool) {
			if rect != (types.Rectf{Min: types.Ptf(-10, -10), Max: types.Ptf(10, 10)}) {
				t.Fatalf("query rect = %v", rect)
			}
			fn(1)
		},
		isMovable: func(int) bool { return true },
		position:  func(int) types.Pointf { return types.Ptf(15, 0) },
		traceRay:  func(types.Pointf, types.Pointf) bool { return true },
		mass:      func(int) float32 { return 1 },
		isMissile: func(int) bool { return false },
		loadForce: func(int) types.Pointf { return types.Pointf{} },
		wake:      func(int) {},
		storeForce: func(int, types.Pointf, types.Pointf) {
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
	if mapPushUnitsAround52E040(&types.Pointf{}, 1, 0, 1, mapPushUnitsAroundHooks52E040[int]{}) {
		t.Fatal("missing dependencies were accepted")
	}
	if new(Server).MapPushUnitsAround52E040(types.Pointf{}, 1, 0, 1, MapPushUnitsAroundRuntime52E040{}) {
		t.Fatal("missing force callback was accepted")
	}
}

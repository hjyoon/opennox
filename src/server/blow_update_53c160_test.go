package server

import (
	"image"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestBlowUpdate53C160DisabledDoesNotScan(t *testing.T) {
	blowUpdateNative53C160(new(Object), blowUpdateDeps53C160{
		indexedDirection: func(int16) image.Point {
			t.Fatal("disabled blow classified its direction")
			return image.Point{}
		},
		eachInRect: func(types.Rectf, func(*Object) bool) {
			t.Fatal("disabled blow scanned the world")
		},
		canInteract: func(*Object, *Object) bool {
			t.Fatal("disabled blow checked interaction")
			return false
		},
		directionVector: func(byte) (float32, float32) {
			t.Fatal("disabled blow loaded a direction vector")
			return 0, 0
		},
	})
}

func TestBlowUpdate53C160FiltersAndAppliesForce(t *testing.T) {
	source := &Object{
		ObjFlags:   object.FlagEnabled,
		PosVec:     types.Ptf(100, 200),
		Direction1: 0,
	}
	valid := &Object{PosVec: types.Ptf(110, 202), ForceVec: types.Ptf(3, 4), Mass: 2}
	destroyed := &Object{PosVec: types.Ptf(110, 202), ObjFlags: object.FlagDestroyed, Mass: 2}
	immobile := &Object{PosVec: types.Ptf(110, 202), ObjClass: object.ClassImmobile, Mass: 2}
	blocked := &Object{PosVec: types.Ptf(110, 202), Mass: 2}
	offAxis := &Object{PosVec: types.Ptf(110, 203), Mass: 2}
	outside := &Object{PosVec: types.Ptf(500, 200), Mass: 2}
	candidates := []*Object{valid, destroyed, immobile, blocked, offAxis, outside}

	var gotRect types.Rectf
	var interactionChecks []*Object
	blowUpdateNative53C160(source, blowUpdateDeps53C160{
		indexedDirection: func(direction int16) image.Point {
			if direction != 0 {
				t.Fatalf("direction = %d, want 0", direction)
			}
			return image.Pt(1, 0)
		},
		eachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			gotRect = rect
			for _, candidate := range candidates {
				if !fn(candidate) {
					t.Fatal("blow scan stopped early")
				}
			}
		},
		canInteract: func(gotSource, candidate *Object) bool {
			if gotSource != source {
				t.Fatalf("interaction source = %p, want %p", gotSource, source)
			}
			interactionChecks = append(interactionChecks, candidate)
			return candidate != blocked
		},
		directionVector: func(direction byte) (float32, float32) {
			if direction != 0 {
				t.Fatalf("vector direction = %d, want 0", direction)
			}
			return 1, 0
		},
	})

	wantRect := types.Rectf{Min: types.Ptf(100, 200), Max: types.Ptf(500, 600)}
	if gotRect != wantRect {
		t.Fatalf("scan rect = %+v, want %+v", gotRect, wantRect)
	}
	if len(interactionChecks) != 2 || interactionChecks[0] != valid || interactionChecks[1] != blocked {
		t.Fatalf("interaction checks = %p, want valid and blocked only", interactionChecks)
	}
	distance := float32(math.Sqrt(104) + float64(blowDistanceBias53C240))
	difference := float64(blowRadius53C160 - distance)
	force := float32(difference * difference * difference * float64(blowForceScale53C240))
	wantX := float32(float64(force)/2 + 3)
	if valid.ForceVec.X != wantX || valid.ForceVec.Y != 4 {
		t.Fatalf("valid force = %+v, want {%v 4}", valid.ForceVec, wantX)
	}
	for name, candidate := range map[string]*Object{
		"destroyed": destroyed,
		"immobile":  immobile,
		"blocked":   blocked,
		"off-axis":  offAxis,
		"outside":   outside,
	} {
		if candidate.ForceVec != (types.Pointf{}) {
			t.Fatalf("%s force = %+v, want zero", name, candidate.ForceVec)
		}
	}
}

func TestBlowUpdateDirectionAllowed53C240MatchesOriginalThresholds(t *testing.T) {
	for _, tc := range []struct {
		name    string
		indexed image.Point
		delta   types.Pointf
		want    bool
	}{
		{name: "east inside cone", indexed: image.Pt(1, 0), delta: types.Ptf(10, 2), want: true},
		{name: "east outside cone", indexed: image.Pt(1, 0), delta: types.Ptf(10, 3)},
		{name: "east wrong sign", indexed: image.Pt(1, 0), delta: types.Ptf(-10, 2)},
		{name: "center accepts", indexed: image.Pt(0, 0), delta: types.Ptf(-10, 20), want: true},
		{name: "original southeast diagonal bounds are empty", indexed: image.Pt(1, 1), delta: types.Ptf(10, 10)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := blowUpdateDirectionAllowed53C240(tc.indexed, tc.delta); got != tc.want {
				t.Fatalf("allowed = %t, want %t", got, tc.want)
			}
		})
	}
}

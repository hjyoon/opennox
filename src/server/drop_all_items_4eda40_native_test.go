package server

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDropAllItemsNative4EDA40ObjectLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantPos := uintptr(56)
	wantShape := uintptr(172)
	wantRadius := uintptr(176)
	wantNext := uintptr(496)
	wantHead := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantPos = 60
		wantShape = 176
		wantRadius = 180
		wantNext = 528
		wantHead = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.Shape", unsafe.Offsetof(Object{}.Shape), wantShape},
		{"Object.Shape.Circle.R", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Circle) + unsafe.Offsetof(Shape{}.Circle.R), wantRadius},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantHead},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"drop ray size", unsafe.Sizeof(dropAllItemsRay4EDA40{}), 16},
		{"drop ray origin", unsafe.Offsetof(dropAllItemsRay4EDA40{}.Origin), 0},
		{"drop ray destination", unsafe.Offsetof(dropAllItemsRay4EDA40{}.Destination), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDropAllItemsNative4EDA40BindsLiveObjectsAndTraceLocal(t *testing.T) {
	item := &Object{}
	item.Shape.Circle.R = 2
	owner := &Object{
		ObjClass:     object.ClassPlayer,
		PosVec:       types.Pointf{X: 10, Y: 20},
		InvFirstItem: item,
	}
	randomLines := make([]int32, 0, 2)
	var traceRay *dropAllItemsRay4EDA40
	var droppedPoint *types.Pointf
	deps := dropAllItemsNativeDeps4EDA40{
		randomFloat: func(min, max float32, source string, line int32) float64 {
			if min != -3 || max != 3 || source != dropAllItemsSource4EDA40 {
				t.Fatalf("random boundary = %v/%v/%q", min, max, source)
			}
			randomLines = append(randomLines, line)
			return 0
		},
		mapTrace: func(ray *dropAllItemsRay4EDA40, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			traceRay = ray
			if outPoint != nil || outGrid != nil || flags != 1 {
				t.Fatalf("trace optional outputs/flags = %p/%p/%d", outPoint, outGrid, flags)
			}
			if ray.Origin != (types.Pointf{X: 10, Y: 20}) || ray.Destination != (types.Pointf{X: 10, Y: 30}) {
				t.Fatalf("trace ray = %+v", *ray)
			}
			ray.Destination = types.Pointf{X: 31, Y: 41}
			return math.MinInt32
		},
		drop: func(gotOwner, gotItem *Object, point *types.Pointf) int32 {
			if gotOwner != owner || gotItem != item {
				t.Fatalf("drop objects = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
			}
			droppedPoint = point
			if *point != (types.Pointf{X: 31, Y: 41}) {
				t.Fatalf("drop point = %+v", *point)
			}
			return -7
		},
	}

	if got := dropAllItemsNative4EDA40(owner, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if len(randomLines) != 2 || randomLines[0] != 823 || randomLines[1] != 824 {
		t.Fatalf("random lines = %v, want [823 824]", randomLines)
	}
	if traceRay == nil || droppedPoint != &traceRay.Destination || droppedPoint == &owner.PosVec {
		t.Fatalf("trace/drop/local point identities = %p/%p/%p", traceRay, droppedPoint, &owner.PosVec)
	}
}

func TestDropAllItemsNative4EDA40FallbackUsesExactLiveOwnerPosition(t *testing.T) {
	item := new(Object)
	owner := &Object{PosVec: types.Pointf{X: 5, Y: 6}, InvFirstItem: item}
	randomCalls := 0
	traceCalls := 0
	dropCalls := 0
	deps := dropAllItemsNativeDeps4EDA40{
		randomFloat: func(float32, float32, string, int32) float64 {
			randomCalls++
			return 0
		},
		mapTrace: func(*dropAllItemsRay4EDA40, *types.Pointf, *image.Point, uint8) int32 {
			traceCalls++
			if traceCalls == 9 {
				owner.PosVec = types.Pointf{X: 50, Y: 60}
			}
			return 0
		},
		drop: func(gotOwner, gotItem *Object, point *types.Pointf) int32 {
			dropCalls++
			if gotOwner != owner || gotItem != item || point != &owner.PosVec || *point != (types.Pointf{X: 50, Y: 60}) {
				t.Fatalf("fallback = %p/%p/%p %+v", gotOwner, gotItem, point, *point)
			}
			return math.MinInt32
		},
	}
	if got := dropAllItemsNative4EDA40(owner, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if randomCalls != 18 || traceCalls != 9 || dropCalls != 1 {
		t.Fatalf("random/trace/drop calls = %d/%d/%d, want 18/9/1", randomCalls, traceCalls, dropCalls)
	}
}

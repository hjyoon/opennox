package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestChestOpenNative4EDF00ObjectAndRayLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubClass := uintptr(12)
	wantFlags := uintptr(16)
	wantPosition := uintptr(56)
	wantShape := uintptr(172)
	wantCircleRadius := uintptr(176)
	wantBoxW := uintptr(184)
	wantBoxH := uintptr(188)
	wantWeight := uintptr(488)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubClass = 16
		wantFlags = 20
		wantPosition = 60
		wantShape = 176
		wantCircleRadius = 180
		wantBoxW = 188
		wantBoxH = 192
		wantWeight = 516
		wantNext = 528
		wantFirst = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.Shape", unsafe.Offsetof(Object{}.Shape), wantShape},
		{"Object.Shape.Circle.R", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Circle) + unsafe.Offsetof(Shape{}.Circle.R), wantCircleRadius},
		{"Object.Shape.Box.W", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Box) + unsafe.Offsetof(ShapeBox{}.W), wantBoxW},
		{"Object.Shape.Box.H", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Box) + unsafe.Offsetof(ShapeBox{}.H), wantBoxH},
		{"Object.Weight", unsafe.Offsetof(Object{}.Weight), wantWeight},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"chest ray size", unsafe.Sizeof(chestOpenRay4EDF00{}), 16},
		{"chest ray origin", unsafe.Offsetof(chestOpenRay4EDF00{}.Origin), 0},
		{"chest ray destination", unsafe.Offsetof(chestOpenRay4EDF00{}.Destination), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestChestShapeExtentNative4EE2A0UsesNamedLiveShapeFields(t *testing.T) {
	obj := new(Object)
	obj.Shape.Kind = ShapeKindCircle
	obj.Shape.Circle.R = 7.5
	if got := chestShapeExtentNative4EE2A0(obj); got != 7.5 {
		t.Fatalf("circle extent = %v, want 7.5", got)
	}

	obj.Shape.Kind = ShapeKindBox
	obj.Shape.Box.W = 8
	obj.Shape.Box.H = 12
	if got := chestShapeExtentNative4EE2A0(obj); got != 6 {
		t.Fatalf("box extent = %v, want 6", got)
	}

	obj.Shape.Kind = ShapeKindCenter
	obj.Shape.Circle.R = 99
	if got := chestShapeExtentNative4EE2A0(obj); math.Float64bits(got) != 0 {
		t.Fatalf("other extent bits = %016x, want positive zero", math.Float64bits(got))
	}
}

func TestChestOpenNative4EDF00BindsPointersLiveFieldsAndServices(t *testing.T) {
	item := &Object{
		Weight:   1,
		ObjClass: object.ClassImmobile,
		ObjFlags: object.Flags(0x80000001),
	}
	chest := &Object{
		ObjSubClass:  0x400,
		PosVec:       types.Pointf{X: 10, Y: 20},
		InvFirstItem: item,
	}
	chest.Shape.Kind = ShapeKindBox
	chest.Shape.Box.W = 4
	chest.Shape.Box.H = 8
	unit := &Object{PosVec: types.Pointf{X: 100, Y: 200}}

	var normalized *types.Pointf
	var traceRay *chestOpenRay4EDF00
	var traceInputs []chestOpenRay4EDF00
	var events []string
	var droppedPoint *types.Pointf
	deps := chestOpenNativeDeps4EDF00{
		normalize: func(point *types.Pointf) {
			events = append(events, "normalize")
			normalized = point
			*point = types.Pointf{X: 1, Y: 0}
		},
		mapTrace: func(ray *chestOpenRay4EDF00, outPoint *types.Pointf, outGrid *image.Point, flags uint8) int32 {
			events = append(events, "trace")
			if outPoint != nil || outGrid != nil || flags != 1 {
				t.Fatalf("trace optional outputs/flags = %p/%p/%d", outPoint, outGrid, flags)
			}
			if traceRay == nil {
				traceRay = ray
			} else if traceRay != ray {
				t.Fatalf("trace ray changed from %p to %p", traceRay, ray)
			}
			traceInputs = append(traceInputs, *ray)
			ray.Origin.X++
			ray.Destination = types.Pointf{X: 999, Y: 999}
			return math.MinInt32
		},
		refresh: func(got *Object) {
			events = append(events, "refresh")
			if got != item || uint32(item.ObjFlags) != 0x80000041 {
				t.Fatalf("refresh object/flags = %p/%08x", got, uint32(item.ObjFlags))
			}
		},
		drop: func(gotChest, gotItem *Object, point *types.Pointf) int32 {
			events = append(events, "drop")
			if gotChest != chest || gotItem != item {
				t.Fatalf("drop objects = %p/%p, want %p/%p", gotChest, gotItem, chest, item)
			}
			droppedPoint = point
			return math.MinInt32
		},
	}

	chestOpenNative4EDF00(chest, unit, deps)
	if normalized == nil || *normalized != (types.Pointf{X: 1, Y: 0}) {
		t.Fatalf("normalized local = %p %+v", normalized, normalized)
	}
	if len(traceInputs) != 3 || traceInputs[0].Origin != (types.Pointf{X: 10, Y: 20}) || traceInputs[1].Origin != (types.Pointf{X: 11, Y: 20}) || traceInputs[2].Origin != (types.Pointf{X: 12, Y: 20}) {
		t.Fatalf("trace inputs = %+v", traceInputs)
	}
	if droppedPoint == nil || *droppedPoint != (types.Pointf{X: 33, Y: -10}) {
		t.Fatalf("dropped point = %p %+v, want local (33,-10)", droppedPoint, droppedPoint)
	}
	if droppedPoint == &chest.PosVec || droppedPoint == &unit.PosVec || uint32(item.ObjFlags) != 0x80000041 {
		t.Fatalf("drop point aliases object or flags wrong: point=%p chest=%p unit=%p flags=%08x", droppedPoint, &chest.PosVec, &unit.PosVec, uint32(item.ObjFlags))
	}
	if want := []string{"normalize", "trace", "trace", "trace", "refresh", "drop"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestChestOpenNative4EDF00TraceFailureUsesLiveUnitAndCachesInventorySuccessor(t *testing.T) {
	second := &Object{Weight: 1}
	first := &Object{Weight: 1, InvNextItem: second}
	chest := &Object{
		ObjSubClass:  0x100,
		PosVec:       types.Pointf{X: 1, Y: 2},
		InvFirstItem: first,
	}
	unit := &Object{PosVec: types.Pointf{X: 3, Y: 4}}
	traceCalls := 0
	var drops []struct {
		item  *Object
		point types.Pointf
	}
	deps := chestOpenNativeDeps4EDF00{
		normalize: func(*types.Pointf) {},
		mapTrace: func(*chestOpenRay4EDF00, *types.Pointf, *image.Point, uint8) int32 {
			traceCalls++
			if traceCalls == 3 {
				unit.PosVec = types.Pointf{X: 30, Y: 40}
				return 0
			}
			return 1
		},
		refresh: func(*Object) {},
		drop: func(_ *Object, item *Object, point *types.Pointf) int32 {
			drops = append(drops, struct {
				item  *Object
				point types.Pointf
			}{item: item, point: *point})
			if item == first {
				first.InvNextItem = nil
			}
			return -1
		},
	}

	chestOpenNative4EDF00(chest, unit, deps)
	if traceCalls != 3 || len(drops) != 2 || drops[0].item != first || drops[1].item != second {
		t.Fatalf("trace/drops = %d/%+v", traceCalls, drops)
	}
	if drops[0].point != (types.Pointf{X: 30, Y: 40}) {
		t.Fatalf("first drop point = %+v, want live unit fallback", drops[0].point)
	}
}

func TestChestOpen4EDF00ServerNilGateDoesNotTouchRuntime(t *testing.T) {
	runtime := ChestOpenRuntime4EDF00{
		NormalizeVector: func(*types.Pointf) { t.Fatal("normalize called") },
		RefreshUnit:     func(*Object) { t.Fatal("refresh called") },
		Dispatch: func(*Object, *Object, *types.Pointf) int32 {
			t.Fatal("dispatch called")
			return 0
		},
	}
	new(Server).ChestOpen4EDF00(nil, new(Object), runtime)
}

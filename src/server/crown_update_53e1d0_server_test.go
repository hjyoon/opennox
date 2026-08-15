package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultCrownUpdateNativeDeps53E1D0() crownUpdateNativeDeps53E1D0 {
	return crownUpdateNativeDeps53E1D0{
		pickup:     func(*Object, *Object, int32, int32) uint32 { return 0 },
		clearOwner: func(*Object) {},
		trace:      func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false },
		move:       func(*Object, types.Pointf) {},
	}
}

func TestCrownUpdate53E1D0NativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectSize := uintptr(780)
	wantFlags := uintptr(16)
	wantPos := uintptr(56)
	wantDirection := uintptr(124)
	wantShape := uintptr(172)
	wantOwner := uintptr(508)
	wantUpdate := uintptr(748)
	if ptrSize == 8 {
		wantObjectSize = 928
		wantFlags = 20
		wantPos = 60
		wantDirection = 128
		wantShape = 176
		wantOwner = 552
		wantUpdate = 872
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection},
		{"Object.Shape", unsafe.Offsetof(Object{}.Shape), wantShape},
		{"Shape.Circle", unsafe.Offsetof(Shape{}.Circle), 4},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestCrownUpdateNative53E1D0PickupUsesNativePointers(t *testing.T) {
	target := &Object{ObjFlags: object.Flags(0x40000000)}
	update := &CrownUpdateData{PickupTarget: target}
	crown := &Object{UpdateData: unsafe.Pointer(update)}
	deps := defaultCrownUpdateNativeDeps53E1D0()
	calls := 0
	deps.pickup = func(who, item *Object, flag1, flag2 int32) uint32 {
		calls++
		if who != target || item != crown || flag1 != 1 || flag2 != 1 {
			t.Fatalf("pickup args = (%p,%p,%d,%d)", who, item, flag1, flag2)
		}
		return 0xf1234567
	}

	crownUpdateNative53E1D0(crown, deps)
	if calls != 1 {
		t.Fatalf("pickup calls = %d, want 1", calls)
	}
}

func TestCrownUpdateNative53E1D0ClearsNativeFallbackAndOwner(t *testing.T) {
	field0 := &Object{ObjFlags: object.Flags(0x20)}
	owner := &Object{ObjFlags: object.FlagDead}
	update := &CrownUpdateData{Field0: field0}
	crown := &Object{UpdateData: unsafe.Pointer(update), ObjOwner: owner}
	deps := defaultCrownUpdateNativeDeps53E1D0()
	deps.clearOwner = func(obj *Object) {
		if obj != crown || update.Field0 != nil {
			t.Fatalf("clear owner obj/field0 = (%p,%p)", obj, update.Field0)
		}
		obj.ObjOwner = nil
	}

	crownUpdateNative53E1D0(crown, deps)
	if update.Field0 != nil || crown.ObjOwner != nil {
		t.Fatalf("fallback state = (%p,%p), want nils", update.Field0, crown.ObjOwner)
	}
}

func TestCrownUpdateNative53E1D0TracesAndMovesUsingDirectionTable(t *testing.T) {
	owner := &Object{
		PosVec:     types.Pointf{X: 30, Y: -20},
		Direction1: 32,
	}
	owner.Shape.Circle.R = 2
	crown := &Object{
		ObjOwner:   owner,
		UpdateData: unsafe.Pointer(&CrownUpdateData{}),
	}
	crown.Shape.Circle.R = 3
	deps := defaultCrownUpdateNativeDeps53E1D0()
	events := make([]string, 0, 2)
	var tracedTo, movedTo types.Pointf
	deps.trace = func(from, to types.Pointf, flags MapTraceFlags) bool {
		if from != owner.PosVec || flags != MapTraceFlags(5) {
			t.Fatalf("trace from/flags = (%+v,%d)", from, flags)
		}
		events = append(events, "trace")
		tracedTo = to
		return true
	}
	deps.move = func(obj *Object, destination types.Pointf) {
		if obj != crown {
			t.Fatalf("move object = %p, want %p", obj, crown)
		}
		events = append(events, "move")
		movedTo = destination
	}

	crownUpdateNative53E1D0(crown, deps)
	cosine, sine := SinCosDir(32)
	distance := float64(15)
	want := types.Pointf{
		X: float32(distance*float64(cosine) + 30),
		Y: float32(distance*float64(sine) - 20),
	}
	if tracedTo != want || movedTo != want {
		t.Fatalf("destinations = (%+v,%+v), want %+v", tracedTo, movedTo, want)
	}
	if !reflect.DeepEqual(events, []string{"trace", "move"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestCrownUpdate53E1D0ServerForwardsPickupRuntime(t *testing.T) {
	s := &Server{}
	target := &Object{}
	crown := &Object{UpdateData: unsafe.Pointer(&CrownUpdateData{PickupTarget: target})}
	calls := 0
	runtime := CrownUpdateRuntime53E1D0{
		Pickup: func(who, item *Object, flag1, flag2 int32) uint32 {
			calls++
			if who != target || item != crown || flag1 != 1 || flag2 != 1 {
				t.Fatalf("pickup args = (%p,%p,%d,%d)", who, item, flag1, flag2)
			}
			return 1
		},
		Move: func(*Object, types.Pointf) {
			t.Fatal("pickup path moved Crown")
		},
	}

	s.CrownUpdate53E1D0(crown, runtime)
	if calls != 1 {
		t.Fatalf("pickup calls = %d, want 1", calls)
	}
}

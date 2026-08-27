package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultInventoryServPlaceNativeDeps4F36F0() inventoryServPlaceNativeDeps4F36F0 {
	return inventoryServPlaceNativeDeps4F36F0{
		itemTypeAllowed: func(uint16) int32 { return 1 },
		callPickup: func(PickupFuncPtr, *Object, *Object, int32, int32) int32 {
			return 1
		},
		defaultPickup:  func(*Object, *Object, int32, int32) int32 { return 1 },
		refreshCollide: func(*Object) {},
		scriptPickup:   func(*ScriptCallback, *Object, *Object) {},
	}
}

func TestInventoryServPlace4F36F0NativeLayoutAndConstants(t *testing.T) {
	checks32 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 780},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 4},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 16},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 490},
		{"Object.Collide", unsafe.Offsetof(Object{}.Collide), 696},
		{"Object.Pickup", unsafe.Offsetof(Object{}.Pickup), 708},
		{"Object.ScriptPickup", unsafe.Offsetof(Object{}.ScriptPickup), 764},
		{"PickupFuncPtr.size", unsafe.Sizeof(PickupFuncPtr{}), 4},
		{"PickupFuncPtr.Ptr", unsafe.Offsetof(PickupFuncPtr{}.Ptr), 0},
		{"ScriptCallback.size", unsafe.Sizeof(ScriptCallback{}), 8},
		{"ScriptCallback.Func", unsafe.Offsetof(ScriptCallback{}.Func), 4},
	}
	checks64 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 8},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 12},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), 20},
		{"Object.CarryCapacity", unsafe.Offsetof(Object{}.CarryCapacity), 518},
		{"Object.Collide", unsafe.Offsetof(Object{}.Collide), 768},
		{"Object.Pickup", unsafe.Offsetof(Object{}.Pickup), 792},
		{"Object.ScriptPickup", unsafe.Offsetof(Object{}.ScriptPickup), 904},
		{"PickupFuncPtr.size", unsafe.Sizeof(PickupFuncPtr{}), 8},
		{"PickupFuncPtr.Ptr", unsafe.Offsetof(PickupFuncPtr{}.Ptr), 0},
		{"ScriptCallback.size", unsafe.Sizeof(ScriptCallback{}), 8},
		{"ScriptCallback.Func", unsafe.Offsetof(ScriptCallback{}.Func), 4},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}

	constants := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"destroyed", uint32(object.FlagDestroyed), uint32(inventoryServPlaceDestroyedFlagLow4F36F0)},
		{"dead", uint32(object.FlagDead), inventoryServPlaceDeadFlag4F36F0},
		{"unit class", uint32(object.MaskUnits), uint32(inventoryServPlaceUnitClassLow4F36F0)},
		{"no collide", uint32(object.FlagNoCollide), inventoryServPlaceNoCollideFlag4F36F0},
	}
	for _, check := range constants {
		if check.got != check.want {
			t.Errorf("%s = %#x, want %#x", check.name, check.got, check.want)
		}
	}
}

func TestInventoryServPlaceNative4F36F0PreservesPointersResultAndLivePostState(t *testing.T) {
	var pickupStorage byte
	var initialCollideStorage byte
	var liveCollideStorage byte
	owner := &Object{
		ObjClass:      object.ClassPlayer | object.Class(0x80000000),
		ObjFlags:      object.FlagMarked,
		CarryCapacity: 0xabcd,
	}
	item := &Object{
		TypeInd:      0xf123,
		ObjFlags:     object.FlagBelow,
		Collide:      unsafe.Pointer(&initialCollideStorage),
		Pickup:       PickupFuncPtr{Ptr: unsafe.Pointer(&pickupStorage)},
		ScriptPickup: ScriptCallback{Func: -1},
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(item)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	events := make([]string, 0, 4)
	deps := defaultInventoryServPlaceNativeDeps4F36F0()
	deps.itemTypeAllowed = func(typeInd uint16) int32 {
		events = append(events, "allowed")
		if typeInd != item.TypeInd {
			t.Fatalf("type = %#x, want %#x", typeInd, item.TypeInd)
		}
		return math.MinInt32
	}
	deps.callPickup = func(
		pickup PickupFuncPtr,
		gotOwner, gotItem *Object,
		arg3, arg4 int32,
	) int32 {
		events = append(events, "pickup")
		if pickup.Ptr != item.Pickup.Ptr || gotOwner != owner || gotItem != item {
			t.Fatalf("pickup/objects = %p/%p/%p", pickup.Ptr, gotOwner, gotItem)
		}
		if arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("args = %d/%d", arg3, arg4)
		}
		item.ObjFlags = object.FlagMarked | object.FlagNoCollide | object.FlagBelow
		item.Collide = unsafe.Pointer(&liveCollideStorage)
		return math.MinInt32
	}
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("custom pickup called DefaultPickup")
		return 0
	}
	deps.refreshCollide = func(gotItem *Object) {
		events = append(events, "refresh")
		if gotItem != item || item.ObjFlags.Has(object.FlagNoCollide) {
			t.Fatalf("refresh item/flags = %p/%#x", gotItem, item.ObjFlags)
		}
		if item.Collide != unsafe.Pointer(&liveCollideStorage) {
			t.Fatalf("live collide = %p", item.Collide)
		}
		item.ScriptPickup.Func = 0x10203040
	}
	deps.scriptPickup = func(callback *ScriptCallback, gotOwner, gotItem *Object) {
		events = append(events, "script")
		if callback != &item.ScriptPickup || gotOwner != owner || gotItem != item {
			t.Fatalf("script callback/objects = %p/%p/%p", callback, gotOwner, gotItem)
		}
		callback.Func = math.MinInt32
	}

	if got := inventoryServPlaceNative4F36F0(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if want := []string{"allowed", "pickup", "refresh", "script"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if want := object.FlagMarked | object.FlagBelow; item.ObjFlags != want {
		t.Fatalf("item flags = %#x, want %#x", item.ObjFlags, want)
	}
	if item.ScriptPickup.Func != -1 {
		t.Fatalf("script Func = %d, want -1", item.ScriptPickup.Func)
	}
}

func TestInventoryServPlaceNative4F36F0NilPickupUsesExactDefault(t *testing.T) {
	owner := &Object{ObjClass: object.ClassMonster, CarryCapacity: 1}
	item := &Object{TypeInd: math.MaxUint16, ScriptPickup: ScriptCallback{Func: -1}}
	deps := defaultInventoryServPlaceNativeDeps4F36F0()
	deps.callPickup = func(PickupFuncPtr, *Object, *Object, int32, int32) int32 {
		t.Fatal("nil pickup called custom callback")
		return 0
	}
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		if gotOwner != owner || gotItem != item || arg3 != -17 || arg4 != -23 {
			t.Fatalf("default = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		return math.MaxInt32
	}
	if got := inventoryServPlaceNative4F36F0(owner, item, -17, -23, deps); got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
}

func TestInventoryServPlaceServerDeps4F36F0RejectMissingAndDisabledTypes(t *testing.T) {
	s := &Server{}
	s.Types.byInd = []*ObjectType{
		nil,
		{allowed: true},
		{allowed: false},
	}
	deps := inventoryServPlaceServerDeps4F36F0(s, InventoryServPlaceRuntime4F36F0{})
	for _, test := range []struct {
		typeInd uint16
		want    int32
	}{
		{0, 0},
		{1, 1},
		{2, 0},
		{3, 0},
		{math.MaxUint16, 0},
	} {
		if got := deps.itemTypeAllowed(test.typeInd); got != test.want {
			t.Errorf("allowed(%d) = %d, want %d", test.typeInd, got, test.want)
		}
	}
}

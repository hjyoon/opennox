package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func defaultPickupAbilityBookNativeDeps4F3CE0() pickupAbilityBookNativeDeps4F3CE0 {
	return pickupAbilityBookNativeDeps4F3CE0{
		gameFlagsCheck: func(uint32) int32 { return 0 },
		useByNetCode:   func(*Object, *Object) {},
		defaultPickup:  func(*Object, *Object, int32, int32) int32 { return 0 },
		audio:          func(uint32, *Object, int32, uint32) {},
	}
}

func TestPickupAbilityBook4F3CE0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantFlags = 20
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"callback result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPickupAbilityBookNative4F3CE0UsesNativePointersAndFourArgs(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32) {
		t.Fatalf("native pointers do not exercise high 64-bit half: owner=%p item=%p", owner, item)
	}

	var events []string
	deps := defaultPickupAbilityBookNativeDeps4F3CE0()
	deps.gameFlagsCheck = func(flags uint32) int32 {
		events = append(events, "flags")
		if flags != 0x1800 {
			t.Fatalf("game flags = %#x, want 0x1800", flags)
		}
		return -1
	}
	deps.useByNetCode = func(gotOwner, gotItem *Object) {
		events = append(events, "use")
		if gotOwner != owner || gotItem != item {
			t.Fatalf("use objects = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
	}
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		return math.MinInt32
	}
	deps.audio = func(sound uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if sound != 826 || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%#x", sound, gotOwner, kind, code)
		}
	}
	if got := pickupAbilityBookNative4F3CE0(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(events, []string{"flags", "use", "default", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestPickupAbilityBookNative4F3CE0UseCanDestroyItem(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	deps := defaultPickupAbilityBookNativeDeps4F3CE0()
	deps.gameFlagsCheck = func(uint32) int32 { return 1 }
	deps.useByNetCode = func(gotOwner, gotItem *Object) {
		if gotOwner != owner || gotItem != item {
			t.Fatalf("use objects = %p/%p", gotOwner, gotItem)
		}
		item.ObjFlags |= 0x20
	}
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("destroyed item reached DefaultPickup")
		return 0
	}
	if got := pickupAbilityBookNative4F3CE0(owner, item, 3, 4, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
}

func TestPickupAbilityBookNative4F3CE0PreservesNilItemFault(t *testing.T) {
	deps := defaultPickupAbilityBookNativeDeps4F3CE0()
	defer func() {
		if recover() == nil {
			t.Fatal("nil item did not preserve the ObjFlags fault")
		}
	}()
	pickupAbilityBookNative4F3CE0(&Object{}, nil, 0, 0, deps)
}

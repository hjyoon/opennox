package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestCrownCollide4EBB50NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
		wantFlags = 20
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestCrownCollideNative4EBB50GuardReturnsFullTargetPointer(t *testing.T) {
	collision := types.Pointf{X: 3.5, Y: -8.25}
	deps := crownCollideNativeDeps4EBB50{
		pickup: func(*Object, *Object, int32, int32) uint32 {
			t.Fatal("guard called CrownPickup")
			return 0
		},
	}
	if got := crownCollideNative4EBB50(nil, nil, &collision, deps); got != 0 {
		t.Fatalf("nil-target result = %#x, want 0", got)
	}

	target := &Object{ObjFlags: object.FlagDestroyed, ObjClass: object.ClassPlayer}
	got := crownCollideNative4EBB50(nil, target, &collision, deps)
	if want := uintptr(target.CObj()); got != want {
		t.Fatalf("guard result = %#x, want full pointer %#x", got, want)
	}
	if collision != (types.Pointf{X: 3.5, Y: -8.25}) {
		t.Fatalf("collision mutated to %+v", collision)
	}
}

func TestCrownCollideNative4EBB50PickupArgumentsAndResult(t *testing.T) {
	crown := &Object{}
	target := &Object{ObjClass: object.ClassPlayer}
	called := 0
	deps := crownCollideNativeDeps4EBB50{
		pickup: func(who, gotCrown *Object, flag1, flag2 int32) uint32 {
			called++
			if who != target || gotCrown != crown || flag1 != 1 || flag2 != 1 {
				t.Fatalf("pickup args = (%p,%p,%d,%d)", who, gotCrown, flag1, flag2)
			}
			return 0xf1234567
		},
	}

	got := crownCollideNative4EBB50(crown, target, nil, deps)
	if want := uintptr(uint32(0xf1234567)); got != want {
		t.Fatalf("pickup result = %#x, want zero-extended %#x", got, want)
	}
	if called != 1 {
		t.Fatalf("pickup calls = %d, want 1", called)
	}
}

func TestCrownCollide4EBB50ServerForwardsRuntime(t *testing.T) {
	s := &Server{}
	crown := &Object{}
	target := &Object{ObjClass: object.ClassPlayer}
	collision := &types.Pointf{X: 1, Y: 2}
	called := false

	got := s.CrownCollide4EBB50(crown, target, collision, CrownCollideRuntime4EBB50{
		Pickup: func(who, gotCrown *Object, flag1, flag2 int32) uint32 {
			called = true
			if who != target || gotCrown != crown || flag1 != 1 || flag2 != 1 {
				t.Fatalf("pickup args = (%p,%p,%d,%d)", who, gotCrown, flag1, flag2)
			}
			return 9
		},
	})
	if got != 9 || !called {
		t.Fatalf("result = %#x, called = %v", got, called)
	}
}

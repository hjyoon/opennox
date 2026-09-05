package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestLifetimeUpdateNative53B8F0Layout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantFlags := uintptr(16)
	wantCreation := uintptr(128)
	wantDeath := uintptr(724)
	wantUpdateData := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantFlags = 20
		wantCreation = 132
		wantDeath = 824
		wantUpdateData = 872
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Field32", unsafe.Offsetof(Object{}.Field32), wantCreation},
		{"Object.Death", unsafe.Offsetof(Object{}.Death), wantDeath},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"LifetimeUpdateData size", unsafe.Sizeof(LifetimeUpdateData53B8F0{}), 4},
		{"LifetimeUpdateData.Duration", unsafe.Offsetof(LifetimeUpdateData53B8F0{}.Duration), 0},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestLifetimeUpdateNative53B8F0KeepsWidePointersAndCallOrder(t *testing.T) {
	data := &LifetimeUpdateData53B8F0{Duration: 30}
	deathToken := new(uint32)
	source := &Object{
		Field32:    999,
		ObjFlags:   object.Flags(0xa5001234),
		Death:      unsafe.Pointer(deathToken),
		UpdateData: unsafe.Pointer(data),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"source":         source.CObj(),
			"update data":    unsafe.Pointer(data),
			"death callback": unsafe.Pointer(deathToken),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above the ABI32 range", name, ptr)
			}
		}
	}

	var events []string
	var called unsafe.Pointer
	lifetimeUpdateNative53B8F0(source, lifetimeUpdateNativeDeps53B8F0{
		frame: func() uint32 {
			events = append(events, "frame")
			// GAME.EXE samples the frame before loading the creation frame.
			source.Field32 = 100
			return 131
		},
		callDeath: func(death unsafe.Pointer, got *Object) {
			events = append(events, "death")
			called = death
			if got != source || got.ObjFlags != object.Flags(0xa5009234) {
				t.Fatalf("death object/flags = (%p, %#x), want (%p, 0xa5009234)", got, got.ObjFlags, source)
			}
		},
		delayedDelete: func(*Object) {
			t.Fatal("death callback path used delayed deletion")
		},
	})
	if !reflect.DeepEqual(events, []string{"frame", "death"}) || called != unsafe.Pointer(deathToken) {
		t.Fatalf("events/callback = (%#v, %p), want ([frame death], %p)", events, called, deathToken)
	}
	runtime.KeepAlive(data)
	runtime.KeepAlive(deathToken)
	runtime.KeepAlive(source)
}

func TestLifetimeUpdateNative53B8F0NilDeathFallback(t *testing.T) {
	data := &LifetimeUpdateData53B8F0{}
	source := &Object{
		ObjFlags:   object.Flags(0x40000001),
		UpdateData: unsafe.Pointer(data),
	}
	var deleted *Object
	lifetimeUpdateNative53B8F0(source, lifetimeUpdateNativeDeps53B8F0{
		frame:     func() uint32 { return 1 },
		callDeath: func(unsafe.Pointer, *Object) { t.Fatal("nil death callback invoked") },
		delayedDelete: func(got *Object) {
			if got.ObjFlags != object.Flags(0x40008001) {
				t.Fatalf("delete flags = %#x, want 0x40008001", got.ObjFlags)
			}
			deleted = got
		},
	})
	if deleted != source {
		t.Fatalf("deleted source = %p, want %p", deleted, source)
	}
	runtime.KeepAlive(data)
}

func TestServerLifetimeUpdate53B8F0BindsFrameAndFallback(t *testing.T) {
	s := new(Server)
	s.SetFrame(51)
	data := &LifetimeUpdateData53B8F0{Duration: 50}
	source := &Object{UpdateData: unsafe.Pointer(data)}
	var deleted *Object
	s.LifetimeUpdate53B8F0(source, LifetimeUpdateRuntime53B8F0{
		DelayedDelete: func(got *Object) { deleted = got },
	})
	if deleted != source || !source.ObjFlags.Has(object.FlagDead) {
		t.Fatalf("deleted/dead = (%p, %t), want (%p, true)", deleted, source.ObjFlags.Has(object.FlagDead), source)
	}
	runtime.KeepAlive(data)
}

package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDieCollide4E99B0NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantDeath := uintptr(724)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
		wantDeath = 824
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Death", unsafe.Offsetof(Object{}.Death), wantDeath},
	} {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDieCollideNative4E99B0OrderCachedDeathAndFlags(t *testing.T) {
	token := new(uint32)
	source := &Object{ObjFlags: object.Flags(0x21), Death: unsafe.Pointer(token)}
	target := &Object{ObjClass: object.ClassPlayer}
	var events []string
	var called unsafe.Pointer
	dieCollideNative4E99B0(source, target, unsafe.Pointer(new(uint32)), dieCollideNativeDeps4E99B0{
		unitsOnSameTeam: func(first, second *Object) int32 {
			events = append(events, "same")
			if first != source || second != target {
				t.Fatalf("same-team args = (%p, %p), want (%p, %p)", first, second, source, target)
			}
			return 0
		},
		callDeath: func(death unsafe.Pointer, obj *Object) {
			events = append(events, "death")
			called = death
			if obj != source || uint32(source.ObjFlags) != 0x8021 {
				t.Fatalf("death obj/flags = (%p, %#x), want (%p, 0x8021)", obj, source.ObjFlags, source)
			}
		},
		delayedDelete: func(*Object) { t.Fatal("death callback path delayed deletion") },
	})
	if !reflect.DeepEqual(events, []string{"same", "death"}) || called != unsafe.Pointer(token) {
		t.Fatalf("events/callback = (%#v, %p), want ([same death], %p)", events, called, token)
	}
}

func TestDieCollideNative4E99B0NilDeathFallback(t *testing.T) {
	source := &Object{ObjFlags: object.Flags(0x40000000)}
	target := &Object{ObjClass: object.ClassMonster}
	var deleted *Object
	dieCollideNative4E99B0(source, target, nil, dieCollideNativeDeps4E99B0{
		unitsOnSameTeam: func(*Object, *Object) int32 { return 0 },
		callDeath:       func(unsafe.Pointer, *Object) { t.Fatal("nil death callback invoked") },
		delayedDelete:   func(obj *Object) { deleted = obj },
	})
	if deleted != source || uint32(source.ObjFlags) != 0x40008000 {
		t.Fatalf("deleted/flags = (%p, %#x), want (%p, 0x40008000)", deleted, source.ObjFlags, source)
	}
}

func TestDieCollideNative4E99B0NilTargetDoesNotReadSource(t *testing.T) {
	dieCollideNative4E99B0(nil, nil, nil, dieCollideNativeDeps4E99B0{
		unitsOnSameTeam: func(*Object, *Object) int32 {
			t.Fatal("nil target called same-team")
			return 0
		},
	})
}

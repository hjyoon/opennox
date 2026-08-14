package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSignCollide4EAB40NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjClass := uintptr(8)
	wantUse := uintptr(732)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjClass = 12
		wantUse = 840
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjClass},
		{"Object.Use", unsafe.Offsetof(Object{}.Use), wantUse},
		{"UseFuncPtr size", unsafe.Sizeof(UseFuncPtr{}), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSignCollideNative4EAB40UsesLowClassAndCachedUse(t *testing.T) {
	var tokenA, tokenB byte
	ptrA := unsafe.Pointer(&tokenA)
	ptrB := unsafe.Pointer(&tokenB)
	source := &Object{Use: UseFuncPtr{Ptr: ptrA}, Field188: 0x11223344}
	target := &Object{ObjClass: object.ClassPlayer | object.Class(0x80000000), Field188: 0x55667788}
	collision := &types.Pointf{X: 3, Y: 4}
	calls := 0
	objUse.Register(ptrA, func(gotTarget, gotSource *Object) bool {
		calls++
		if gotTarget != target || gotSource != source {
			t.Fatalf("Use args = %p/%p", gotTarget, gotSource)
		}
		source.Use.Ptr = ptrB
		target.Field188++
		return true
	})
	objUse.Register(ptrB, func(*Object, *Object) bool {
		t.Fatal("replacement Use callback reached")
		return false
	})

	signCollideNative4EAB40(source, target, collision)
	if calls != 1 || source.Use.Ptr != ptrB || source.Field188 != 0x11223344 || target.Field188 != 0x55667789 || collision.X != 3 || collision.Y != 4 {
		t.Fatalf("calls/state = %d/%#x/%#x/%+v", calls, source.Field188, target.Field188, collision)
	}
}

func TestSignCollide4EAB40ServerBindingAndTargetGuards(t *testing.T) {
	s := &Server{}
	s.SignCollide4EAB40(nil, nil, nil)
	s.SignCollide4EAB40(nil, &Object{ObjClass: object.ClassImmobile}, nil)
}

func TestSignCollideNative4EAB40NilSourceFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
	}()
	signCollideNative4EAB40(nil, &Object{ObjClass: object.ClassPlayer}, nil)
}

func TestSignCollideNative4EAB40NilUseFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Use callback did not fault")
		}
	}()
	signCollideNative4EAB40(&Object{}, &Object{ObjClass: object.ClassPlayer}, nil)
}

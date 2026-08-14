package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestFistCollide4EADF0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantZ := uintptr(104)
	wantDamage := uintptr(716)
	wantUpdate := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantZ = 108
		wantDamage = 808
		wantUpdate = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ZVal", unsafe.Offsetof(Object{}.ZVal), wantZ},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"FistUpdateData size", unsafe.Sizeof(FistUpdateData{}), 4},
		{"FistUpdateData.Damage", unsafe.Offsetof(FistUpdateData{}.Damage), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFistCollideNative4EADF0OrderAndLiveDamageCallback(t *testing.T) {
	data := &FistUpdateData{Damage: 0x71234567}
	parent := &Object{}
	oldMarker := uint32(1)
	newMarker := uint32(2)
	oldFn := unsafe.Pointer(&oldMarker)
	newFn := unsafe.Pointer(&newMarker)
	source := &Object{UpdateData: unsafe.Pointer(data)}
	target := &Object{Damage: oldFn}
	var events []string
	var gotFn unsafe.Pointer
	var gotArgs [3]*Object
	var gotDamage int32
	var gotType object.DamageType

	fistCollideNative4EADF0(source, target, nil, fistCollideNativeDeps4EADF0{
		findParentPlayer: func(got *Object) *Object {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p, want %p", got, source)
			}
			data.Damage = -9
			source.UpdateData = unsafe.Pointer(&FistUpdateData{Damage: -10})
			target.Damage = newFn
			return parent
		},
		callTargetDamage: func(
			fn unsafe.Pointer,
			gotTarget, gotParent, gotSource *Object,
			damage int32,
			damageType object.DamageType,
		) int32 {
			events = append(events, "damage")
			gotFn = fn
			gotArgs = [3]*Object{gotTarget, gotParent, gotSource}
			gotDamage = damage
			gotType = damageType
			return -0x1234567
		},
	})

	if want := []string{"parent", "damage"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if gotFn != newFn || gotFn == oldFn {
		t.Fatalf("Damage callback = %p, want live %p", gotFn, newFn)
	}
	if want := [3]*Object{target, parent, source}; gotArgs != want {
		t.Fatalf("Damage args = %v, want %v", gotArgs, want)
	}
	if gotDamage != 0x71234567 || gotType != object.DamageType(fistCollideDamageType4EADF0) {
		t.Fatalf("Damage value/type = %#x/%d", gotDamage, gotType)
	}
}

func TestFistCollideNative4EADF0NilTargetDoesNotDereferenceNilUpdate(t *testing.T) {
	source := &Object{ZVal: math.Float32frombits(0x7fc12345)}
	fistCollideNative4EADF0(source, nil, nil, fistCollideNativeDeps4EADF0{
		findParentPlayer: func(*Object) *Object {
			t.Fatal("parent lookup for nil target")
			return nil
		},
		callTargetDamage: func(unsafe.Pointer, *Object, *Object, *Object, int32, object.DamageType) int32 {
			t.Fatal("Damage call for nil target")
			return 0
		},
	})
}

func TestFistCollideNative4EADF0PositiveHeightSkipsNilUpdate(t *testing.T) {
	fistCollideNative4EADF0(&Object{ZVal: 1}, &Object{}, nil, fistCollideNativeDeps4EADF0{
		findParentPlayer: func(*Object) *Object {
			t.Fatal("parent lookup above ground")
			return nil
		},
		callTargetDamage: func(unsafe.Pointer, *Object, *Object, *Object, int32, object.DamageType) int32 {
			t.Fatal("Damage call above ground")
			return 0
		},
	})
}

func TestFistCollideNative4EADF0NilSourceFaultsAtHeight(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		fistCollideNative4EADF0(nil, &Object{}, nil, fistCollideNativeDeps4EADF0{})
	}()
	if recovered == nil {
		t.Fatal("nil source did not fault")
	}
}

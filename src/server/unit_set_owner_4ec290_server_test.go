package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestUnitSetOwnerNative4EC290Layout(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantOwner := uintptr(508)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
		wantFlags = 20
		wantOwner = 552
		wantNextOwned = 560
		wantFirstOwned = 568
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitSetOwnerNative4EC290UsesNativeFieldsAndCallbackOrder(t *testing.T) {
	oldFirst := &Object{}
	terminal := &Object{Field129: oldFirst}
	destroyed := &Object{ObjFlags: object.FlagDestroyed, ObjOwner: terminal}
	obj := &Object{ObjClass: object.ClassMonster}
	events := make([]string, 0, 2)
	unitSetOwnerNative4EC290(destroyed, obj, unitSetOwnerNativeDeps4EC290{
		clearOwner: func(got *Object) {
			events = append(events, "clear")
			if got != obj {
				t.Fatalf("clear object = %p", got)
			}
		},
		resetMonster: func(got *Object) {
			events = append(events, "reset")
			got.ObjClass = 0
		},
		markUnitUpdate: func(*Object) {
			t.Fatal("unit update used the class cached before reset")
		},
	})
	if !reflect.DeepEqual(events, []string{"clear", "reset"}) {
		t.Fatalf("events = %v", events)
	}
	if obj.ObjOwner != terminal || obj.Field128 != oldFirst || terminal.Field129 != obj {
		t.Fatalf("ownership = owner %p next %p first %p", obj.ObjOwner, obj.Field128, terminal.Field129)
	}
}

func TestUnitSetOwner4EC290ServerBindingReplacesOwnedListEntry(t *testing.T) {
	s := &Server{}
	oldTail := &Object{}
	oldOwner := &Object{Field129: nil}
	obj := &Object{ObjOwner: oldOwner, Field128: oldTail}
	oldOwner.Field129 = obj
	newFirst := &Object{}
	newOwner := &Object{Field129: newFirst}

	s.ObjSetOwner(newOwner, obj)
	if oldOwner.Field129 != oldTail {
		t.Fatalf("old owner first = %p, want %p", oldOwner.Field129, oldTail)
	}
	if obj.ObjOwner != newOwner || obj.Field128 != newFirst || newOwner.Field129 != obj {
		t.Fatalf("new ownership = owner %p next %p first %p", obj.ObjOwner, obj.Field128, newOwner.Field129)
	}
}

func TestUnitSetOwner4EC290ServerBindingSkipsDestroyedOwners(t *testing.T) {
	s := &Server{}
	terminal := &Object{}
	destroyed := &Object{ObjFlags: object.FlagDestroyed, ObjOwner: terminal}
	obj := &Object{}

	s.ObjSetOwner(destroyed, obj)
	if obj.ObjOwner != terminal || terminal.Field129 != obj {
		t.Fatalf("resolved ownership = owner %p first %p", obj.ObjOwner, terminal.Field129)
	}
	s.ObjSetOwner(nil, nil)
}

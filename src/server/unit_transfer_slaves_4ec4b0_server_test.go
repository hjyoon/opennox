package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestUnitTransferSlavesNative4EC4B0Layout(t *testing.T) {
	wantSize := uintptr(780)
	wantOwner := uintptr(508)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
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

func TestUnitTransferSlavesNative4EC4B0UsesLiveOwnerAndCachedNext(t *testing.T) {
	ownerA := &Object{}
	ownerB := &Object{}
	replacement := &Object{}
	first := &Object{}
	second := &Object{}
	source := &Object{ObjOwner: ownerA, Field129: first}
	first.Field128 = second
	gotOwners := make([]*Object, 0, 2)
	gotChildren := make([]*Object, 0, 2)

	unitTransferSlavesNative4EC4B0(source, func(owner, child *Object) {
		gotOwners = append(gotOwners, owner)
		gotChildren = append(gotChildren, child)
		if child == first {
			child.Field128 = replacement
			source.ObjOwner = ownerB
		}
	})

	if !reflect.DeepEqual(gotOwners, []*Object{ownerA, ownerB}) {
		t.Fatalf("owners = %v, want [%p %p]", gotOwners, ownerA, ownerB)
	}
	if !reflect.DeepEqual(gotChildren, []*Object{first, second}) {
		t.Fatalf("children = %v, want [%p %p]", gotChildren, first, second)
	}
	if source.Field129 != first {
		t.Fatalf("source first-owned was explicitly changed to %p", source.Field129)
	}
}

func TestUnitTransferSlaves4EC4B0ServerBindingTransfersOwnedList(t *testing.T) {
	s := &Server{}
	grandparent := &Object{}
	source := &Object{ObjOwner: grandparent}
	grandparent.Field129 = source
	first := &Object{ObjOwner: source}
	second := &Object{ObjOwner: source}
	source.Field129 = first
	first.Field128 = second

	s.ObjTransferSlaves(source)

	if source.Field129 != nil || source.ObjOwner != grandparent {
		t.Fatalf("source ownership = owner %p first %p", source.ObjOwner, source.Field129)
	}
	if grandparent.Field129 != second || second.ObjOwner != grandparent || second.Field128 != first {
		t.Fatalf("second ownership = owner %p next %p grandparent first %p", second.ObjOwner, second.Field128, grandparent.Field129)
	}
	if first.ObjOwner != grandparent || first.Field128 != source {
		t.Fatalf("first ownership = owner %p next %p", first.ObjOwner, first.Field128)
	}

	s.ObjTransferSlaves(nil)
	s.ObjTransferSlaves(&Object{})
}

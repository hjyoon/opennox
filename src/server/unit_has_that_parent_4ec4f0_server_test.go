package server

import (
	"testing"
	"time"
	"unsafe"
)

func TestUnitHasThatParentNative4EC4F0Layout(t *testing.T) {
	wantSize := uintptr(780)
	wantOwner := uintptr(508)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantOwner = 552
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitHasThatParentNative4EC4F0RejectsNilBeforeCycleTraversal(t *testing.T) {
	obj := &Object{}
	obj.ObjOwner = obj
	done := make(chan bool, 1)
	go func() {
		done <- obj.HasOwner(nil)
	}()
	select {
	case got := <-done:
		if got {
			t.Fatal("self-cycle matched a nil owner")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("nil owner traversed a self-cycle")
	}

	var nilObject *Object
	if nilObject.HasOwner(obj) {
		t.Fatal("nil object matched a non-nil owner")
	}
}

func TestUnitHasThatParentNative4EC4F0IncludesSelfAndOwnerChain(t *testing.T) {
	owner := &Object{}
	middle := &Object{ObjOwner: owner}
	obj := &Object{ObjOwner: middle}
	missing := &Object{}

	if !obj.HasOwner(obj) {
		t.Fatal("object did not match itself")
	}
	if !obj.HasOwner(middle) || !obj.HasOwner(owner) {
		t.Fatal("reachable owner did not match")
	}
	if obj.HasOwner(missing) {
		t.Fatal("unreachable owner unexpectedly matched")
	}
}

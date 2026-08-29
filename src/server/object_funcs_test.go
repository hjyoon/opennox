package server

import (
	"testing"
	"unsafe"
)

func TestObjectPickupHandlerReturnsExactRegistration(t *testing.T) {
	const name = "ObjectPickupHandlerTest"
	if _, ok := pickupFuncs[name]; ok {
		t.Fatalf("test pickup handler %q is already registered", name)
	}
	var storage byte
	want := unsafe.Pointer(&storage)
	pickupFuncs[name] = want
	t.Cleanup(func() { delete(pickupFuncs, name) })

	got, ok := ObjectPickupHandler(name)
	if !ok || got.Ptr != want {
		t.Fatalf("ObjectPickupHandler(%q) = %p/%t, want %p/true", name, got.Ptr, ok, want)
	}
	if got, ok := ObjectPickupHandler(name + "Missing"); ok || got.Ptr != nil {
		t.Fatalf("missing ObjectPickupHandler = %p/%t, want nil/false", got.Ptr, ok)
	}
}

func TestObjectDeathHandlerReturnsExactRegistration(t *testing.T) {
	const name = "ObjectDeathHandlerTest"
	if _, ok := deathFuncs[name]; ok {
		t.Fatalf("test death handler %q is already registered", name)
	}
	var storage byte
	wantPtr := unsafe.Pointer(&storage)
	wantSize := uintptr(132)
	deathFuncs[name] = objectDefFunc{Func: wantPtr, DataSize: wantSize}
	t.Cleanup(func() { delete(deathFuncs, name) })

	gotPtr, gotSize, ok := ObjectDeathHandler(name)
	if !ok || gotPtr != wantPtr || gotSize != wantSize {
		t.Fatalf("ObjectDeathHandler(%q) = %p/%d/%t, want %p/%d/true", name, gotPtr, gotSize, ok, wantPtr, wantSize)
	}
	if gotPtr, gotSize, ok := ObjectDeathHandler(name + "Missing"); ok || gotPtr != nil || gotSize != 0 {
		t.Fatalf("missing ObjectDeathHandler = %p/%d/%t, want nil/0/false", gotPtr, gotSize, ok)
	}
}

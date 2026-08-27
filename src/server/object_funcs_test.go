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

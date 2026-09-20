package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestDeathBallFragmentUpdate53D220DispatchStaysInGo(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("DeathBallFragmentUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("DeathBallFragmentUpdate registration = %p/%d/%t, want nonnil/0/true", callback, size, ok)
	}

	oldCall := deathBallFragmentUpdateCall53D220
	t.Cleanup(func() { deathBallFragmentUpdateCall53D220 = oldCall })

	owner := new(server.Object)
	fragment := &server.Object{ObjOwner: owner}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(fragment)) <= 0xffffffff {
		t.Fatalf("fragment pointer = %p, want above 4 GiB", fragment)
	}
	var called *server.Object
	deathBallFragmentUpdateCall53D220 = func(got *server.Object) {
		called = got
	}

	// ObjOwner deliberately leaves a live Go pointer in the object. The old
	// dispatcher would send this pointer graph through an indirect C call.
	server.CallObjectUpdate(callback, fragment)
	if called != fragment {
		t.Fatalf("DeathBallFragmentUpdate called with %p, want %p", called, fragment)
	}
}

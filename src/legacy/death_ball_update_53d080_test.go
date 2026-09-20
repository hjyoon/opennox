package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestDeathBallUpdate53D080DispatchStaysInGo(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("DeathBallUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("DeathBallUpdate registration = %p/%d/%t, want nonnil/0/true", callback, size, ok)
	}

	oldUpdate := Nox_xxx_updateDeathBall_53D080
	t.Cleanup(func() { Nox_xxx_updateDeathBall_53D080 = oldUpdate })

	owner := new(server.Object)
	ball := &server.Object{ObjOwner: owner}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ball)) <= 0xffffffff {
		t.Fatalf("death ball pointer = %p, want above 4 GiB", ball)
	}
	var called *server.Object
	Nox_xxx_updateDeathBall_53D080 = func(got *server.Object) {
		called = got
	}

	// ObjOwner deliberately leaves a live Go pointer in the object. Regressing
	// to the legacy C trampoline would send that pointer graph through cgo.
	server.CallObjectUpdate(callback, ball)
	if called != ball {
		t.Fatalf("DeathBallUpdate called with %p, want %p", called, ball)
	}
}

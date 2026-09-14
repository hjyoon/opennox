package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestUndeadKillerUpdate53E190RegistrationUsesNativeObject(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("UndeadKillerUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("UndeadKillerUpdate registration = %p/%d/%t, want nonnil/0/true", callback, size, ok)
	}
	obj := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= uintptr(^uint32(0)) {
		t.Fatalf("object pointer = %p, want above 4 GiB", obj)
	}
	previous := undeadKillerUpdateCall53E190
	t.Cleanup(func() { undeadKillerUpdateCall53E190 = previous })
	var calls int
	undeadKillerUpdateCall53E190 = func(got *server.Object) {
		calls++
		if got != obj {
			t.Errorf("object = %p, want %p", got, obj)
		}
	}
	server.CallObjectUpdate(callback, obj)
	if calls != 1 {
		t.Fatalf("native calls = %d, want 1", calls)
	}
}

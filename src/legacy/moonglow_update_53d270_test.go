package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMoonglowUpdate53D270RegistrationUsesNativeObject(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("MoonglowUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("MoonglowUpdate registration = %p/%d/%t, want nonnil/0/true", callback, size, ok)
	}
	visual := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(visual)) <= 0xffffffff {
		t.Fatalf("visual pointer = %p, want above 4 GiB", visual)
	}
	oldCall := moonglowUpdateCall53D270
	t.Cleanup(func() { moonglowUpdateCall53D270 = oldCall })
	calls := 0
	moonglowUpdateCall53D270 = func(got *server.Object) {
		calls++
		if got != visual {
			t.Errorf("visual = %p, want %p", got, visual)
		}
	}
	server.CallObjectUpdate(callback, visual)
	if calls != 1 {
		t.Fatalf("native calls = %d, want 1", calls)
	}
}

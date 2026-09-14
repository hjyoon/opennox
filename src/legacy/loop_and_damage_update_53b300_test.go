package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestLoopAndDamageUpdate53B300RegistrationUsesNativeObject(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("LoopAndDamageUpdate")
	if !ok || callback == nil || size != 16 {
		t.Fatalf("LoopAndDamageUpdate registration = %p/%d/%t, want nonnil/16/true", callback, size, ok)
	}
	obj := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= 0xffffffff {
		t.Fatalf("object pointer = %p, want above 4 GiB", obj)
	}
	server.CallObjectUpdate(callback, obj)
	if obj.ObjFlags != 0x40 {
		t.Fatalf("native flags = %#x, want 0x40", obj.ObjFlags)
	}
}

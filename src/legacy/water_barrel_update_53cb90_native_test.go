package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestWaterBarrelUpdate53CB90RegistrationAndCGoRoundTrip(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("WaterBarrelUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("WaterBarrelUpdate registration = %p/%d/%t, want nonnil/0/true", callback, size, ok)
	}
	source := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want above ABI32 range", source)
	}
	oldCall := waterBarrelUpdateCall53CB90
	t.Cleanup(func() { waterBarrelUpdateCall53CB90 = oldCall })
	var calls int
	waterBarrelUpdateCall53CB90 = func(got *server.Object) {
		calls++
		if got != source {
			t.Fatalf("WaterBarrelUpdate source = %p, want %p", got, source)
		}
	}
	server.CallObjectUpdate(callback, source)
	waterBarrelUpdateExportCall53CB90(source)
	if calls != 2 {
		t.Fatalf("native/CGo calls = %d, want 2", calls)
	}
	runtime.KeepAlive(source)
}

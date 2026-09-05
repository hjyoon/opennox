package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestLifetimeUpdateRegistrationKeepsTypedNativeContract53B8F0(t *testing.T) {
	wantCallback := Get_nox_xxx_updateLifetime_53B8F0()
	wantSize := unsafe.Sizeof(server.LifetimeUpdateData53B8F0{})
	gotCallback, gotSize, ok := server.ObjectUpdateHandler("LifetimeUpdate")
	if !ok || gotCallback == nil || gotCallback != wantCallback || gotSize != wantSize {
		t.Fatalf("LifetimeUpdate registration = %p/%d/%t, want %p/%d/true", gotCallback, gotSize, ok, wantCallback, wantSize)
	}
}

func TestLifetimeUpdateExportKeepsNativePointerWidth53B8F0(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	source := new(server.Object)
	if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	oldCall := lifetimeUpdateCall53B8F0
	t.Cleanup(func() { lifetimeUpdateCall53B8F0 = oldCall })
	var calls int
	lifetimeUpdateCall53B8F0 = func(got *server.Object) {
		calls++
		if got != source {
			t.Fatalf("LifetimeUpdate source = %p, want %p", got, source)
		}
	}

	lifetimeUpdateExportCall53B8F0(source)
	if calls != 1 {
		t.Fatalf("export calls = %d, want 1", calls)
	}
	runtime.KeepAlive(source)
}

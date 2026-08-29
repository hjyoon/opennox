package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestExpireUpdateExportKeepsNativePointerWidth53DB00(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	source := new(server.Object)
	if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	oldCall := expireUpdateCall53DB00
	t.Cleanup(func() { expireUpdateCall53DB00 = oldCall })
	var calls int
	expireUpdateCall53DB00 = func(got *server.Object) {
		calls++
		if got != source {
			t.Fatalf("ExpireUpdate source = %p, want %p", got, source)
		}
	}

	expireUpdateExportCall53DB00(source)
	if calls != 1 {
		t.Fatalf("export calls = %d, want 1", calls)
	}
	runtime.KeepAlive(source)
}

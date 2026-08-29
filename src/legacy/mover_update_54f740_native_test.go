package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMoverUpdateExportKeepsNativePointerWidth54F740(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	source := new(server.Object)
	if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	oldCall := moverUpdateCall54F740
	t.Cleanup(func() { moverUpdateCall54F740 = oldCall })
	var calls int
	moverUpdateCall54F740 = func(got *server.Object) {
		calls++
		if got != source {
			t.Fatalf("MoverUpdate source = %p, want %p", got, source)
		}
	}

	moverUpdateExportCall54F740(source)
	if calls != 1 {
		t.Fatalf("export calls = %d, want 1", calls)
	}
	runtime.KeepAlive(source)
}

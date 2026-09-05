package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitIdleExport515820PreservesNativePointer(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	unit := new(server.Object)
	if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}

	oldCall := unitIdleCall515820
	t.Cleanup(func() { unitIdleCall515820 = oldCall })
	got := make([]*server.Object, 0, 2)
	unitIdleCall515820 = func(unit *server.Object) {
		got = append(got, unit)
	}

	unitIdleExportCall515820(unit)
	unitIdleExportCall515820(nil)
	if len(got) != 2 {
		t.Fatalf("export calls = %d, want 2", len(got))
	}
	if got[0] != unit {
		t.Fatalf("unit = %p, want %p", got[0], unit)
	}
	if got[1] != nil {
		t.Fatalf("null unit = %p, want nil", got[1])
	}
	runtime.KeepAlive(unit)
}

func TestUnitIdleGoWrapper515820UsesNativeAdapter(t *testing.T) {
	unit := new(server.Object)
	oldCall := unitIdleCall515820
	t.Cleanup(func() { unitIdleCall515820 = oldCall })

	var got *server.Object
	unitIdleCall515820 = func(unit *server.Object) { got = unit }
	Nox_xxx_unitIdle_515820(unit)
	if got != unit {
		t.Fatalf("unit = %p, want %p", got, unit)
	}
}

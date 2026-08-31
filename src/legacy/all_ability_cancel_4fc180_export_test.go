package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestAllAbilityCancelExport4FC180PreservesNativePointer(t *testing.T) {
	unit := new(server.Object)
	var (
		gotUnit *server.Object
		calls   int
	)
	old := Nox_xxx_playerCancelAbils_4FC180
	Nox_xxx_playerCancelAbils_4FC180 = func(unit *server.Object) {
		gotUnit = unit
		calls++
	}
	t.Cleanup(func() { Nox_xxx_playerCancelAbils_4FC180 = old })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	allAbilityCancelExportCall4FC180(unit)
	if calls != 1 || gotUnit != unit {
		t.Fatalf("C export = calls %d, unit %p; want 1, %p", calls, gotUnit, unit)
	}
	runtime.KeepAlive(unit)
}

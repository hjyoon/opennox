package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestSingleAbilityResetExport4FC0B0PreservesNativePointerAndSignedInt32(t *testing.T) {
	unit := new(server.Object)
	var (
		gotUnit    *server.Object
		gotAbility int32
		calls      int
	)
	old := Sub_4FC0B0
	Sub_4FC0B0 = func(unit *server.Object, ability int32) {
		gotUnit = unit
		gotAbility = ability
		calls++
	}
	t.Cleanup(func() { Sub_4FC0B0 = old })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	singleAbilityResetExportCall4FC0B0(unit, math.MinInt32)
	if calls != 1 || gotUnit != unit || gotAbility != math.MinInt32 {
		t.Fatalf("C export = calls %d, unit %p, ability %d; want 1, %p, %d", calls, gotUnit, gotAbility, unit, int32(math.MinInt32))
	}
	runtime.KeepAlive(unit)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestActiveAbilityDisableExport4FC300PreservesNativePointerAndSignedInt32(t *testing.T) {
	unit := new(server.Object)
	var (
		gotUnit    *server.Object
		gotAbility int32
		calls      int
	)
	old := Sub_4FC300
	Sub_4FC300 = func(unit *server.Object, ability int32) {
		gotUnit = unit
		gotAbility = ability
		calls++
	}
	t.Cleanup(func() { Sub_4FC300 = old })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	activeAbilityDisableExportCall4FC300(unit, math.MinInt32)
	if calls != 1 || gotUnit != unit || gotAbility != math.MinInt32 {
		t.Fatalf("C export = calls %d, unit %p, ability %d; want 1, %p, %d", calls, gotUnit, gotAbility, unit, int32(math.MinInt32))
	}
	runtime.KeepAlive(unit)
}

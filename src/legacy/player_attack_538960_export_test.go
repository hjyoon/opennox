package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackExport538960KeepsNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	// Keep this object in Go-managed memory. On 64-bit Linux the legacy C heap
	// may sit below 4 GiB because the test binary is non-PIE, while the Go heap
	// still exercises the native high half on every supported 64-bit host. All
	// pointer fields remain nil, so passing it for the duration of this CGo call
	// does not retain a Go pointer in C memory.
	unit := new(server.Object)
	if pointer := uintptr(unsafe.Pointer(unit)); pointer <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}

	// An object without player update data is deliberately used here. The
	// native entry rejects it safely; the decompiled ABI32 body would truncate
	// the pointer before reading the original +748 update-data slot.
	if got := Nox_xxx_playerAttack_538960(unit); got != 0 {
		t.Fatalf("player attack result = %d, want 0 for missing update data", got)
	}
	runtime.KeepAlive(unit)
}

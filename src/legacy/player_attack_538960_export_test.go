package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackExport538960KeepsNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	unit, freeUnit := alloc.New(server.Object{})
	defer freeUnit()
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

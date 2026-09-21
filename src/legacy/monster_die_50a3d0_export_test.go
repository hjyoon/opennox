package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDieExport50A3D0PreservesNativePointer(t *testing.T) {
	unit := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	var (
		got   *server.Object
		calls int
	)
	old := monsterDieExportImpl50A3D0
	monsterDieExportImpl50A3D0 = func(unit *server.Object) int32 {
		got = unit
		calls++
		return 0x12345678
	}
	t.Cleanup(func() { monsterDieExportImpl50A3D0 = old })

	if result := monsterDieExportCall50A3D0(unit); result != 0x12345678 {
		t.Fatalf("C export result = %#x, want 0x12345678", result)
	}
	if calls != 1 || got != unit {
		t.Fatalf("C export = calls %d, unit %p; want 1, %p", calls, got, unit)
	}
	runtime.KeepAlive(unit)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitBuffTimerExport4FF550PreservesNativePointerSignedDwordAndResult(t *testing.T) {
	type call struct {
		unit *server.Object
		buff int32
	}
	var calls []call
	old := unitBuffTimerExportImpl4FF550
	unitBuffTimerExportImpl4FF550 = func(unit *server.Object, buff int32) uint32 {
		calls = append(calls, call{unit: unit, buff: buff})
		if buff == math.MinInt32 {
			return math.MaxUint16
		}
		return 0x8000
	}
	t.Cleanup(func() { unitBuffTimerExportImpl4FF550 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got := unitBuffTimerExportCall4FF550(unit, math.MinInt32); got != math.MaxUint16 {
		t.Fatalf("minimum-dword result = %#08x, want 0x0000ffff", got)
	}
	if got := unitBuffTimerExportCall4FF550(nil, math.MaxInt32); got != 0x8000 {
		t.Fatalf("maximum-dword result = %#08x, want 0x00008000", got)
	}
	want := []call{
		{unit: unit, buff: math.MinInt32},
		{unit: nil, buff: math.MaxInt32},
	}
	if len(calls) != len(want) {
		t.Fatalf("export calls = %d, want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("export call %d = %+v, want %+v", i, calls[i], want[i])
		}
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffTimerExport4FF550CallsNativeBinding(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	unit.BuffsDur[0] = 0x8000
	unit.BuffsDur[31] = math.MaxUint16

	old := unitBuffTimerExportImpl4FF550
	unitBuffTimerExportImpl4FF550 = unitBuffTimerLegacy4FF550
	t.Cleanup(func() { unitBuffTimerExportImpl4FF550 = old })

	if got := unitBuffTimerExportCall4FF550(unit, 0); got != 0x8000 {
		t.Fatalf("slot 0 result = %#08x, want 0x00008000", got)
	}
	if got := unitBuffTimerExportCall4FF550(unit, 31); got != math.MaxUint16 {
		t.Fatalf("slot 31 result = %#08x, want 0x0000ffff", got)
	}
	runtime.KeepAlive(unit)
}

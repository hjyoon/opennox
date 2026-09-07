package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitBuffPowerExport4FF570PreservesNativePointerSignedDwordAndByte(t *testing.T) {
	type call struct {
		unit *server.Object
		buff int32
	}
	var calls []call
	old := unitBuffPowerExportImpl4FF570
	unitBuffPowerExportImpl4FF570 = func(unit *server.Object, buff int32) uint8 {
		calls = append(calls, call{unit: unit, buff: buff})
		if buff == math.MinInt32 {
			return math.MaxUint8
		}
		return 0x80
	}
	t.Cleanup(func() { unitBuffPowerExportImpl4FF570 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got := unitBuffPowerExportCall4FF570(unit, math.MinInt32); got != math.MaxUint8 {
		t.Fatalf("minimum-dword result = %#02x, want 0xff", got)
	}
	if got := unitBuffPowerExportCall4FF570(nil, math.MaxInt32); got != 0x80 {
		t.Fatalf("maximum-dword result = %#02x, want 0x80", got)
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

func TestUnitBuffPowerExport4FF570CallsNativeBinding(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	unit.BuffsPower[0] = 0x80
	unit.BuffsPower[31] = math.MaxUint8

	old := unitBuffPowerExportImpl4FF570
	unitBuffPowerExportImpl4FF570 = unitBuffPowerLegacy4FF570
	t.Cleanup(func() { unitBuffPowerExportImpl4FF570 = old })

	if got := unitBuffPowerExportCall4FF570(unit, 0); got != 0x80 {
		t.Fatalf("slot 0 result = %#02x, want 0x80", got)
	}
	if got := unitBuffPowerExportCall4FF570(unit, 31); got != math.MaxUint8 {
		t.Fatalf("slot 31 result = %#02x, want 0xff", got)
	}
	runtime.KeepAlive(unit)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitBuffClearExport4FF580PreservesNativePointer(t *testing.T) {
	var calls []*server.Object
	old := unitBuffClearExportImpl4FF580
	unitBuffClearExportImpl4FF580 = func(unit *server.Object) {
		calls = append(calls, unit)
	}
	t.Cleanup(func() { unitBuffClearExportImpl4FF580 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unitBuffClearExportCall4FF580(unit)
	unitBuffClearExportCall4FF580(nil)
	if len(calls) != 2 || calls[0] != unit || calls[1] != nil {
		t.Fatalf("export calls = %#v, want {%p, nil}", calls, unit)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffClearExport4FF580CallsNativeBinding(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassClientPersist
	unit.Buffs = math.MaxUint32
	for i := range unit.BuffsDur {
		unit.BuffsDur[i] = math.MaxUint16
		unit.BuffsPower[i] = math.MaxUint8
	}

	old := unitBuffClearExportImpl4FF580
	unitBuffClearExportImpl4FF580 = unitBuffClearLegacy4FF580
	t.Cleanup(func() { unitBuffClearExportImpl4FF580 = old })

	unitBuffClearExportCall4FF580(unit)
	if unit.Buffs != 0 || unit.BuffsDur != [32]uint16{} || unit.BuffsPower != [32]uint8{} {
		t.Fatalf("export did not clear native buff state: flags=%#x durations=%#v powers=%#v", unit.Buffs, unit.BuffsDur, unit.BuffsPower)
	}
	runtime.KeepAlive(unit)
}

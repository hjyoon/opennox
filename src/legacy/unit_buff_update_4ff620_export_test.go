package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type unitBuffUpdateLegacyServer4FF620 struct {
	Server
	srv         *server.Server
	incremented *server.Object
}

func (s *unitBuffUpdateLegacyServer4FF620) S() *server.Server {
	return s.srv
}

func (s *unitBuffUpdateLegacyServer4FF620) PlayerIncrementElimDeath4D8D40(unit *server.Object) {
	s.incremented = unit
}

func TestUnitBuffUpdateExport4FF620PreservesNativePointer(t *testing.T) {
	var calls []*server.Object
	old := unitBuffUpdateExportImpl4FF620
	unitBuffUpdateExportImpl4FF620 = func(unit *server.Object) {
		calls = append(calls, unit)
	}
	t.Cleanup(func() { unitBuffUpdateExportImpl4FF620 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unitBuffUpdateExportCall4FF620(unit)
	unitBuffUpdateExportCall4FF620(nil)
	want := []*server.Object{unit, nil}
	if len(calls) != len(want) {
		t.Fatalf("export calls = %d, want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("export call %d = %p, want %p", i, calls[i], want[i])
		}
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffUpdateExport4FF620CallsNativeBinding(t *testing.T) {
	srv := new(server.Server)
	bridge := &unitBuffUpdateLegacyServer4FF620{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldImpl := unitBuffUpdateExportImpl4FF620
	unitBuffUpdateExportImpl4FF620 = unitBuffUpdateLegacy4FF620
	t.Cleanup(func() { unitBuffUpdateExportImpl4FF620 = oldImpl })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	unit.Buffs = uint32(1) << 9
	unit.SpeedCur = 8
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unitBuffUpdateExportCall4FF620(unit)
	if unit.SpeedCur != 10 {
		t.Fatalf("native buff-9 speed = %g, want 10", unit.SpeedCur)
	}
	runtime.KeepAlive(unit)
}

func TestUnitBuffUpdateRuntime4FF620PreservesIncrementPointer(t *testing.T) {
	bridge := &unitBuffUpdateLegacyServer4FF620{}
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	unitBuffUpdateRuntime4FF620(bridge).IncrementElimDeath(unit)
	if bridge.incremented != unit {
		t.Fatalf("increment unit = %p, want full native identity %p", bridge.incremented, unit)
	}
	runtime.KeepAlive(unit)
}

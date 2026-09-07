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

type spellBuffOffLegacyServer4FF5B0 struct {
	Server
	srv *server.Server
}

func (s *spellBuffOffLegacyServer4FF5B0) S() *server.Server {
	return s.srv
}

func TestSpellBuffOffExport4FF5B0PreservesNativePointerSignedDwordAndResult(t *testing.T) {
	type call struct {
		unit *server.Object
		buff int32
	}
	var calls []call
	old := spellBuffOffExportImpl4FF5B0
	spellBuffOffExportImpl4FF5B0 = func(unit *server.Object, buff int32) int32 {
		calls = append(calls, call{unit: unit, buff: buff})
		if buff == math.MinInt32 {
			return math.MinInt32
		}
		return math.MaxInt32
	}
	t.Cleanup(func() { spellBuffOffExportImpl4FF5B0 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got := spellBuffOffExportCall4FF5B0(unit, math.MinInt32); got != math.MinInt32 {
		t.Fatalf("minimum-dword result = %#08x, want 0x80000000", uint32(got))
	}
	if got := spellBuffOffExportCall4FF5B0(nil, math.MaxInt32); got != math.MaxInt32 {
		t.Fatalf("maximum-dword result = %#08x, want %#08x", uint32(got), uint32(math.MaxInt32))
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

func TestSpellBuffOffExport4FF5B0CallsNativeBinding(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &spellBuffOffLegacyServer4FF5B0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldImpl := spellBuffOffExportImpl4FF5B0
	spellBuffOffExportImpl4FF5B0 = spellBuffOffLegacy4FF5B0
	t.Cleanup(func() { spellBuffOffExportImpl4FF5B0 = oldImpl })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassClientPersist
	unit.Buffs = uint32(1) << server.ENCHANT_CROWN
	unit.BuffsDur[server.ENCHANT_CROWN] = math.MaxUint16
	unit.BuffsPower[server.ENCHANT_CROWN] = math.MaxUint8
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got := spellBuffOffExportCall4FF5B0(unit, int32(server.ENCHANT_CROWN)); got != 0 {
		t.Fatalf("active result = %#08x, want zero", uint32(got))
	}
	if unit.Buffs != 0 || unit.BuffsDur[server.ENCHANT_CROWN] != 0 || unit.BuffsPower[server.ENCHANT_CROWN] != 0 {
		t.Fatalf("export did not clear native buff state: flags=%#x duration=%#x power=%#x", unit.Buffs, unit.BuffsDur[server.ENCHANT_CROWN], unit.BuffsPower[server.ENCHANT_CROWN])
	}
	if unit.Field38 != math.MaxUint32 {
		t.Fatalf("sync marker = %#08x, want %#08x", unit.Field38, uint32(math.MaxUint32))
	}
	runtime.KeepAlive(unit)
}

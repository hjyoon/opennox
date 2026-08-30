package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type playerSpellLegacyServer4FB2A0 struct {
	Server
	got *server.Object
}

func (s *playerSpellLegacyServer4FB2A0) PlayerSpell(unit *server.Object) {
	s.got = unit
}

func TestPlayerSpellExport4FB2A0PreservesNativePointer(t *testing.T) {
	fake := new(playerSpellLegacyServer4FB2A0)
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	playerSpellExportCall4FB2A0(unit)
	if fake.got != unit {
		t.Fatalf("export unit = %p, want %p", fake.got, unit)
	}
	runtime.KeepAlive(unit)
}

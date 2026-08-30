package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type playerCantCastSpellLegacyServer4FD150 struct {
	Server
	srv *server.Server
}

func (s *playerCantCastSpellLegacyServer4FD150) S() *server.Server {
	return s.srv
}

func TestPlayerCantCastSpellExport4FD150PreservesNativePointer(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerCantCastSpellLegacyServer4FD150{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := &server.Object{Buffs: uint32(1) << server.ENCHANT_ANTI_MAGIC}
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got := playerCantCastSpellExportCall4FD150(unit, 90, 1); got != 14 {
		t.Fatalf("anti-magic result = %d, want 14", got)
	}
	unit.Buffs = 0
	if got := playerCantCastSpellExportCall4FD150(unit, 90, 1); got != 0 {
		t.Fatalf("ordinary spell result = %d, want 0", got)
	}
	runtime.KeepAlive(unit)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

type spellPrecheckLegacyServer4FD0E0 struct {
	Server
	srv *server.Server
}

func (s *spellPrecheckLegacyServer4FD0E0) S() *server.Server {
	return s.srv
}

func TestSpellPrecheckExport4FD0E0PreservesNativePointerAndScalarWidths(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &spellPrecheckLegacyServer4FD0E0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	// An undefined spell reaches the owner-chain read through the native
	// pointer, then returns at the enablement gate before loading ObjClass.
	const spellID = int32(-0x1234567)
	if got := spellPrecheckExportCall4FD0E0(unit, spellID); got != 10 {
		t.Fatalf("CGo export result = %d, want 10", got)
	}
	if got := Sub_4FD0E0(unit, spell.ID(spellID)); got != 10 {
		t.Fatalf("direct legacy result = %d, want 10", got)
	}
	runtime.KeepAlive(unit)
}

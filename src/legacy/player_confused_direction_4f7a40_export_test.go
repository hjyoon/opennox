package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type playerConfusedDirectionLegacyServer4F7A40 struct {
	Server
	srv *server.Server
}

func (s *playerConfusedDirectionLegacyServer4F7A40) S() *server.Server {
	return s.srv
}

func TestPlayerConfusedDirectionExport4F7A40PreservesNativePointer(t *testing.T) {
	srv := new(server.Server)
	srv.SetFrame(^uint32(0))
	oldGetServer := GetServer
	GetServer = func() Server { return &playerConfusedDirectionLegacyServer4F7A40{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	unit.Direction2 = server.Dir16(0xffff)
	unit.NetCode = 1
	unit.BuffsPower[server.ENCHANT_CONFUSED] = 2

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4GiB", unit)
	}
	got := playerConfusedDirectionExportCall4F7A40(unit)
	want := srv.PlayerConfusedDirection4F7A40(unit)
	if got != want || got != 205 {
		t.Fatalf("export direction = %d, want %d (205)", got, want)
	}
	runtime.KeepAlive(unit)
}

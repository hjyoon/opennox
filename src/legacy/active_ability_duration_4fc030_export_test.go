package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type activeAbilityDurationLegacyServer4FC030 struct {
	Server
	srv *server.Server
}

func (s *activeAbilityDurationLegacyServer4FC030) S() *server.Server {
	return s.srv
}

func TestActiveAbilityDurationExport4FC030PreservesNativePointerAndInt32Result(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(0)
	unit := new(server.Object)
	srv.Abils.SetExecHead(&server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(-7), Frame: 0x80000000, Active: 0,
	})

	oldGetServer := GetServer
	GetServer = func() Server {
		return &activeAbilityDurationLegacyServer4FC030{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	if got, want := activeAbilityDurationExportCall4FC030(unit, -7), int32(math.MinInt32); got != want {
		t.Fatalf("C export duration = %d (%08x), want %d (%08x)", got, uint32(got), want, uint32(want))
	}
	if got := activeAbilityDurationExportCall4FC030(unit, int32(server.AbilityWarcry)); got != -1 {
		t.Fatalf("C export miss = %d, want -1", got)
	}
	runtime.KeepAlive(unit)
}

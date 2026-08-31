package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type activeAbilityDeadlineLegacyServer4FC070 struct {
	Server
	srv *server.Server
}

func (s *activeAbilityDeadlineLegacyServer4FC070) S() *server.Server {
	return s.srv
}

func TestActiveAbilityDeadlineExport4FC070PreservesNativePointerAndInt32Delta(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1)
	unit := new(server.Object)
	record := &server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(-7), Frame: 99, Active: 0,
	}
	srv.Abils.SetExecHead(record)

	oldGetServer := GetServer
	GetServer = func() Server {
		return &activeAbilityDeadlineLegacyServer4FC070{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	activeAbilityDeadlineExportCall4FC070(unit, -7, math.MinInt32)
	if got, want := record.Frame, uint32(0x80000001); got != want {
		t.Fatalf("C export deadline = %08x, want %08x", got, want)
	}
	if record.Active != 0 {
		t.Fatal("C export no longer covers an inactive matching record")
	}
	runtime.KeepAlive(unit)
}

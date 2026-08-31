package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type activeAbilityMembershipLegacyServer4FC250 struct {
	Server
	srv *server.Server
}

func (s *activeAbilityMembershipLegacyServer4FC250) S() *server.Server {
	return s.srv
}

func TestActiveAbilityMembershipExport4FC250PreservesNativePointerAndInt32Ability(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	unit := &server.Object{ObjClass: object.ClassPlayer}
	record := &server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(math.MinInt32), Frame: math.MaxUint32, Active: 0,
	}
	srv.Abils.SetExecHead(record)

	oldGetServer := GetServer
	GetServer = func() Server {
		return &activeAbilityMembershipLegacyServer4FC250{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	if got := activeAbilityMembershipExportCall4FC250(unit, math.MinInt32); got != 1 {
		t.Fatalf("C export match = %d, want canonical 1", got)
	}
	if got := activeAbilityMembershipExportCall4FC250(unit, math.MaxInt32); got != 0 {
		t.Fatalf("C export miss = %d, want canonical 0", got)
	}
	if record.Unit != unit || record.Abil != server.Ability(math.MinInt32) ||
		record.Frame != math.MaxUint32 || record.Active != 0 {
		t.Fatal("C export membership query mutated its inactive matching record")
	}
	runtime.KeepAlive(unit)
}

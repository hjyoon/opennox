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

type activeAbilityValueLegacyServer4FC3E0 struct {
	Server
	srv *server.Server
}

func (s *activeAbilityValueLegacyServer4FC3E0) S() *server.Server {
	return s.srv
}

func TestActiveAbilityValueExport4FC3E0PreservesNativePointerInt32AbilityAndRawValue(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	unit := &server.Object{ObjClass: object.ClassPlayer}
	record := &server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(math.MinInt32), Frame: math.MaxUint32, Active: 0x89abcdef,
	}
	srv.Abils.SetExecHead(record)

	oldGetServer := GetServer
	GetServer = func() Server {
		return &activeAbilityValueLegacyServer4FC3E0{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	if got := activeAbilityValueExportCall4FC3E0(unit, math.MinInt32); uint32(got) != 0x89abcdef {
		t.Fatalf("C export active bits = %08x, want 89abcdef", uint32(got))
	}
	if got := activeAbilityValueExportCall4FC3E0(unit, math.MaxInt32); got != 0 {
		t.Fatalf("C export miss = %d, want 0", got)
	}
	record.Active = 0
	if got := activeAbilityValueExportCall4FC3E0(unit, math.MinInt32); got != 0 {
		t.Fatalf("C export inactive match = %d, want 0", got)
	}
	if record.Unit != unit || record.Abil != server.Ability(math.MinInt32) ||
		record.Frame != math.MaxUint32 || record.Active != 0 {
		t.Fatal("C export value query mutated its matching record")
	}
	runtime.KeepAlive(unit)
}

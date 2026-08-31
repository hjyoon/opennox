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

type activeAbilityDeactivateLegacyServer4FC440 struct {
	Server
	srv *server.Server
}

func (s *activeAbilityDeactivateLegacyServer4FC440) S() *server.Server {
	return s.srv
}

func TestActiveAbilityDeactivateExport4FC440PreservesNativePointerSignedInt32AndFirstMatch(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	unit := &server.Object{ObjClass: object.ClassPlayer}
	other := new(server.Object)
	duplicate := &server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(math.MinInt32), Frame: 19, Active: 0x55667788,
	}
	match := &server.ExecAbilityClass{
		Unit: unit, Abil: server.Ability(math.MinInt32), Frame: math.MaxUint32,
		Active: 0x89abcdef, Next: duplicate,
	}
	head := &server.ExecAbilityClass{
		Unit: other, Abil: server.Ability(math.MinInt32), Frame: 77,
		Active: math.MaxUint32, Next: match,
	}
	match.Prev = head
	duplicate.Prev = match
	srv.Abils.SetExecHead(head)

	oldGetServer := GetServer
	GetServer = func() Server {
		return &activeAbilityDeactivateLegacyServer4FC440{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	activeAbilityDeactivateExportCall4FC440(unit, math.MinInt32)
	if match.Active != 0 || duplicate.Active != 0x55667788 || head.Active != math.MaxUint32 {
		t.Fatalf("C export Active values = %08x/%08x/%08x, want 00000000/55667788/ffffffff",
			match.Active, duplicate.Active, head.Active)
	}
	if srv.Abils.ExecHead() != head || head.Next != match || match.Next != duplicate ||
		match.Prev != head || duplicate.Prev != match || match.Frame != math.MaxUint32 {
		t.Fatal("C export mutated list topology or deadline")
	}

	activeAbilityDeactivateExportCall4FC440(unit, math.MaxInt32)
	if duplicate.Active != 0x55667788 || head.Active != math.MaxUint32 {
		t.Fatal("C export signed-ability miss mutated Active")
	}
	runtime.KeepAlive(unit)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerCancelSpellsLegacyServer4FEAE0 struct {
	Server
	srv *server.Server
}

func (s *playerCancelSpellsLegacyServer4FEAE0) S() *server.Server {
	return s.srv
}

func usePlayerCancelSpellsLegacyServer4FEAE0(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &playerCancelSpellsLegacyServer4FEAE0{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func requirePlayerCancelSpellsLegacyPointers4FEAE0(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestPlayerCancelSpellsExport4FEAE0PreservesNativePointers(t *testing.T) {
	srv := usePlayerCancelSpellsLegacyServer4FEAE0(t)
	caster, freeCaster := alloc.New(server.Object{})
	other, freeOther := alloc.New(server.Object{})
	recordA, freeA := alloc.New(server.DurSpell{})
	recordB, freeB := alloc.New(server.DurSpell{})
	recordC, freeC := alloc.New(server.DurSpell{})
	t.Cleanup(freeCaster)
	t.Cleanup(freeOther)
	t.Cleanup(freeA)
	t.Cleanup(freeB)
	t.Cleanup(freeC)

	recordA.Caster16 = other
	recordA.Flags88 = 0x12345620
	recordA.Next = recordB
	recordB.Caster16 = caster
	recordB.Flags88 = 0x89abcdee
	recordB.Next = recordC
	recordC.Caster16 = caster
	recordC.Flags88 = 0xfedcba40
	srv.Spells.Dur.List = recordA
	requirePlayerCancelSpellsLegacyPointers4FEAE0(t,
		unsafe.Pointer(caster), unsafe.Pointer(other),
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	got := playerCancelSpellsExportCall4FEAE0(caster)
	if got != 0 || recordA.Flags88 != 0x12345620 || recordB.Flags88 != 0x89abcdef || recordC.Flags88 != 0xfedcba41 {
		t.Fatalf("result/flags = %d/%#x/%#x/%#x, want 0/0x12345620/0x89abcdef/0xfedcba41",
			got, recordA.Flags88, recordB.Flags88, recordC.Flags88)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
}

func TestPlayerCancelSpellsExport4FEAE0MatchesNilCaster(t *testing.T) {
	srv := usePlayerCancelSpellsLegacyServer4FEAE0(t)
	record, freeRecord := alloc.New(server.DurSpell{})
	t.Cleanup(freeRecord)
	record.Flags88 = 0x87654320
	srv.Spells.Dur.List = record
	requirePlayerCancelSpellsLegacyPointers4FEAE0(t, unsafe.Pointer(record))

	if got := playerCancelSpellsExportCall4FEAE0(nil); got != 0 || record.Flags88 != 0x87654321 {
		t.Fatalf("result/flags = %d/%#x, want 0/0x87654321", got, record.Flags88)
	}
	runtime.KeepAlive(record)
}

package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellCancelDurSpellLegacyServer4FEB10 struct {
	Server
	srv *server.Server
}

func (s *spellCancelDurSpellLegacyServer4FEB10) S() *server.Server {
	return s.srv
}

func useSpellCancelDurSpellLegacyServer4FEB10(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellCancelDurSpellLegacyServer4FEB10{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func requireSpellCancelDurSpellLegacyPointers4FEB10(t *testing.T, values ...unsafe.Pointer) {
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

func TestSpellCancelDurSpellExport4FEB10PreservesNativePointersAndSummonRange(t *testing.T) {
	srv := useSpellCancelDurSpellLegacyServer4FEB10(t)
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

	recordA.Spell = 67
	recordA.Caster16 = caster
	recordA.Flags88 = 0x12345620
	recordA.Next = recordB
	recordB.Spell = 75
	recordB.Caster16 = other
	recordB.Flags88 = 0x89abcdee
	recordB.Next = recordC
	recordC.Spell = 114
	recordC.Caster16 = caster
	recordC.Flags88 = 0xfedcba40
	srv.Spells.Dur.List = recordA
	requireSpellCancelDurSpellLegacyPointers4FEB10(t,
		unsafe.Pointer(caster), unsafe.Pointer(other),
		unsafe.Pointer(recordA), unsafe.Pointer(recordB), unsafe.Pointer(recordC),
	)

	spellCancelDurSpellExportCall4FEB10(75, caster)

	if recordA.Flags88 != 0x12345620 || recordB.Flags88 != 0x89abcdee || recordC.Flags88 != 0xfedcba41 {
		t.Fatalf("flags = %#x/%#x/%#x, want 0x12345620/0x89abcdee/0xfedcba41",
			recordA.Flags88, recordB.Flags88, recordC.Flags88)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
	runtime.KeepAlive(recordC)
}

func TestSpellCancelDurSpellExport4FEB10PreservesSignedDword(t *testing.T) {
	srv := useSpellCancelDurSpellLegacyServer4FEB10(t)
	caster, freeCaster := alloc.New(server.Object{})
	record, freeRecord := alloc.New(server.DurSpell{})
	t.Cleanup(freeCaster)
	t.Cleanup(freeRecord)
	record.Spell = uint32(0x80000000)
	record.Caster16 = caster
	record.Flags88 = 0x76543210
	srv.Spells.Dur.List = record
	requireSpellCancelDurSpellLegacyPointers4FEB10(t, unsafe.Pointer(caster), unsafe.Pointer(record))

	spellCancelDurSpellExportCall4FEB10(math.MinInt32, caster)

	if record.Flags88 != 0x76543211 {
		t.Fatalf("flags = %#x, want 0x76543211", record.Flags88)
	}
	runtime.KeepAlive(caster)
	runtime.KeepAlive(record)
}

func TestSpellCancelDurSpellExport4FEB10MatchesNilCasterAndEmpty(t *testing.T) {
	srv := useSpellCancelDurSpellLegacyServer4FEB10(t)
	record, freeRecord := alloc.New(server.DurSpell{})
	t.Cleanup(freeRecord)
	record.Spell = 67
	record.Flags88 = 0x87654320
	srv.Spells.Dur.List = record
	requireSpellCancelDurSpellLegacyPointers4FEB10(t, unsafe.Pointer(record))

	spellCancelDurSpellExportCall4FEB10(67, nil)
	if record.Flags88 != 0x87654321 {
		t.Fatalf("flags = %#x, want 0x87654321", record.Flags88)
	}
	srv.Spells.Dur.List = nil
	spellCancelDurSpellExportCall4FEB10(math.MaxInt32, nil)
	runtime.KeepAlive(record)
}

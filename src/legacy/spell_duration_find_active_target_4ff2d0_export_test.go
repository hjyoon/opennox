package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellDurationFindActiveTargetLegacyServer4FF2D0 struct {
	Server
	srv *server.Server
}

func (s *spellDurationFindActiveTargetLegacyServer4FF2D0) S() *server.Server {
	return s.srv
}

func useSpellDurationFindActiveTargetLegacyServer4FF2D0(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationFindActiveTargetLegacyServer4FF2D0{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })
	return srv
}

func TestSpellDurationFindActiveTargetExport4FF2D0PreservesNativePointers(t *testing.T) {
	srv := useSpellDurationFindActiveTargetLegacyServer4FF2D0(t)
	target, freeTarget := alloc.New(server.Object{})
	other, freeOther := alloc.New(server.Object{})
	recordA, freeA := alloc.New(server.DurSpell{})
	recordB, freeB := alloc.New(server.DurSpell{})
	t.Cleanup(freeTarget)
	t.Cleanup(freeOther)
	t.Cleanup(freeA)
	t.Cleanup(freeB)

	recordA.Flags88 = 0xffffff01
	recordA.Spell = 0x80000000
	recordA.Target48 = target
	recordA.Next = recordB
	recordB.Flags88 = 0xabcdef00
	recordB.Spell = 0x80000000
	recordB.Target48 = target
	srv.Spells.Dur.List = recordA

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"target": unsafe.Pointer(target),
			"other":  unsafe.Pointer(other),
			"record": unsafe.Pointer(recordB),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	if got := spellDurationFindActiveTargetExportCall4FF2D0(math.MinInt32, target); got != unsafe.Pointer(recordB) {
		t.Fatalf("CGo result = %p, want exact record %p", got, recordB)
	}

	recordB.Target48 = other
	if got := spellDurationFindActiveTargetExportCall4FF2D0(math.MinInt32, target); got != nil {
		t.Fatalf("CGo result after target replacement = %p, want nil", got)
	}

	recordA.Flags88 = 0xffffff00
	recordA.Target48 = target
	if got := spellDurationFindActiveTargetExportCall4FF2D0(math.MinInt32, target); got != unsafe.Pointer(recordA) {
		t.Fatalf("CGo result after active-head replacement = %p, want %p", got, recordA)
	}

	srv.Spells.Dur.List = nil
	if got := spellDurationFindActiveTargetExportCall4FF2D0(math.MinInt32, target); got != nil {
		t.Fatalf("empty-list CGo result = %p, want nil", got)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(other)
	runtime.KeepAlive(recordA)
	runtime.KeepAlive(recordB)
}

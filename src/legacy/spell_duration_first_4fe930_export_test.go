package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellDurationFirstLegacyServer4FE930 struct {
	Server
	srv *server.Server
}

func (s *spellDurationFirstLegacyServer4FE930) S() *server.Server {
	return s.srv
}

func useSpellDurationFirstLegacyServer4FE930(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationFirstLegacyServer4FE930{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func TestSpellDurationFirstExport4FE930UsesLiveNativeHead(t *testing.T) {
	srv := useSpellDurationFirstLegacyServer4FE930(t)
	first, freeFirst := alloc.New(server.DurSpell{ID: 0x1234})
	second, freeSecond := alloc.New(server.DurSpell{ID: 0xabcd})
	t.Cleanup(freeFirst)
	t.Cleanup(freeSecond)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"first":  unsafe.Pointer(first),
			"second": unsafe.Pointer(second),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	srv.Spells.Dur.List = first
	if got := spellDurationFirstExportCall4FE930(); got != unsafe.Pointer(first) {
		t.Fatalf("first CGo result = %p, want %p", got, first)
	}
	srv.Spells.Dur.List = second
	if got := spellDurationFirstExportCall4FE930(); got != unsafe.Pointer(second) {
		t.Fatalf("second CGo result = %p, want live replacement %p", got, second)
	}
	srv.Spells.Dur.List = nil
	if got := spellDurationFirstExportCall4FE930(); got != nil {
		t.Fatalf("nil-list CGo result = %p, want nil", got)
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}

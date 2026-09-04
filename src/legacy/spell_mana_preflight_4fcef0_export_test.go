package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type spellManaPreflightLegacyServer4FCEF0 struct {
	Server
	srv *server.Server
}

func (s *spellManaPreflightLegacyServer4FCEF0) S() *server.Server {
	return s.srv
}

func TestSpellManaPreflightExport4FCEF0PreservesNativePointersAndCount(t *testing.T) {
	wasGodMode := noxflags.HasEngine(noxflags.EngineGodMode)
	noxflags.UnsetEngine(noxflags.EngineGodMode)
	t.Cleanup(func() {
		if wasGodMode {
			noxflags.SetEngine(noxflags.EngineGodMode)
		} else {
			noxflags.UnsetEngine(noxflags.EngineGodMode)
		}
	})

	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &spellManaPreflightLegacyServer4FCEF0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	sequence := [2]int32{0, 0}
	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(&sequence[0])
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(&sequence[0])) <= math.MaxUint32 {
			t.Fatalf("sequence pointer = %p, want native address above 4 GiB", &sequence[0])
		}
	}

	if got := spellManaPreflightExportCall4FCEF0(unit, &sequence[0], 2); got != 1 {
		t.Fatalf("positive-count result = %d, want 1", got)
	}
	if got := spellManaPreflightExportCall4FCEF0(unit, &sequence[0], math.MinInt32); got != 1 {
		t.Fatalf("INT32_MIN count result = %d, want 1", got)
	}
	if got := spellManaPreflightExportCall4FCEF0(unit, &sequence[0], 0); got != 0 {
		t.Fatalf("zero-count result = %d, want 0", got)
	}
	if got := spellManaPreflightExportCall4FCEF0(unit, nil, 1); got != 0 {
		t.Fatalf("nil-sequence result = %d, want 0", got)
	}
	if got := spellManaPreflightExportCall4FCEF0(nil, &sequence[0], 1); got != 0 {
		t.Fatalf("nil-unit result = %d, want 0", got)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(sequence)
}

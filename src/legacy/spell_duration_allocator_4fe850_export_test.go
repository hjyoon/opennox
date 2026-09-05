package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type spellDurationAllocatorLegacyServer4FE850 struct {
	Server
	srv *server.Server
}

func (s *spellDurationAllocatorLegacyServer4FE850) S() *server.Server {
	return s.srv
}

func TestSpellDurationAllocatorExport4FE850UsesLiveNativeServer(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationAllocatorLegacyServer4FE850{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
		srv.Spells.Dur.Free()
	})

	list := &server.DurSpell{ID: 0x1234}
	srv.Spells.Dur.List = list
	if got := spellDurationAllocatorExportCall4FE850(); got != 1 {
		t.Fatalf("CGo export result = %d, want canonical 1", got)
	}
	if srv.Spells.Dur.List != list {
		t.Fatalf("duration list = %p, want preserved %p", srv.Spells.Dur.List, list)
	}
	if record := srv.Spells.Dur.NewRaw(); record == nil {
		t.Fatal("CGo export did not install a live native allocation class")
	} else if record.ID != 1 {
		t.Fatalf("first duration record ID = %d, want 1", record.ID)
	}
}

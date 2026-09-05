package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type spellDurationCleanupLegacyServer4FE880 struct {
	Server
	srv *server.Server
}

func (s *spellDurationCleanupLegacyServer4FE880) S() *server.Server {
	return s.srv
}

func TestSpellDurationCleanupExport4FE880UsesLiveNativeServer(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationCleanupLegacyServer4FE880{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})

	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	record := srv.Spells.Dur.NewRaw()
	if record == nil {
		t.Fatal("native allocation returned nil")
	}
	srv.Spells.Dur.List = record

	spellDurationCleanupExportCall4FE880()

	if srv.Spells.Dur.List != nil {
		t.Fatalf("duration list = %p, want nil", srv.Spells.Dur.List)
	}
}

func TestSpellDurationCleanupExport4FE880AcceptsNilAllocator(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationCleanupLegacyServer4FE880{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	srv.Spells.Dur.List = &server.DurSpell{ID: 0x1234}

	spellDurationCleanupExportCall4FE880()

	if srv.Spells.Dur.List != nil {
		t.Fatalf("duration list = %p, want nil", srv.Spells.Dur.List)
	}
}

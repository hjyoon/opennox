package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellDurationSelectiveCleanupLegacyServer4FE8A0 struct {
	Server
	srv *server.Server
}

func (s *spellDurationSelectiveCleanupLegacyServer4FE8A0) S() *server.Server {
	return s.srv
}

func useSpellDurationSelectiveCleanupLegacyServer4FE8A0(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationSelectiveCleanupLegacyServer4FE8A0{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func TestSpellDurationSelectiveCleanupExport4FE8A0ZeroModeUsesLiveNativeServer(t *testing.T) {
	srv := useSpellDurationSelectiveCleanupLegacyServer4FE8A0(t)
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(srv.Spells.Dur.SpellFreeDurations4FE880)
	record := srv.Spells.Dur.NewRaw()
	if record == nil {
		t.Fatal("native duration allocation returned nil")
	}
	srv.Spells.Dur.List = record

	spellDurationSelectiveCleanupExportCall4FE8A0(0)

	if srv.Spells.Dur.List != nil {
		t.Fatalf("duration list = %p, want nil", srv.Spells.Dur.List)
	}
	next := srv.Spells.Dur.NewRaw()
	if next == nil || next.ID != 2 {
		var id uint16
		if next != nil {
			id = next.ID
		}
		t.Fatalf("post-reset allocation = %p ID %d, want nonnil ID 2", next, id)
	}
}

func TestSpellDurationSelectiveCleanupExport4FE8A0NonzeroModePreservesNativePointers(t *testing.T) {
	srv := useSpellDurationSelectiveCleanupLegacyServer4FE8A0(t)
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(srv.Spells.Dur.SpellFreeDurations4FE880)

	player, freePlayer := alloc.New(server.Object{})
	monster, freeMonster := alloc.New(server.Object{})
	t.Cleanup(freePlayer)
	t.Cleanup(freeMonster)
	player.ObjClass = object.ClassPlayer | object.Class(0x80000000)
	monster.ObjClass = object.ClassMonster | object.Class(0x40000000)

	keep := srv.Spells.Dur.NewRaw()
	remove := srv.Spells.Dur.NewRaw()
	if keep == nil || remove == nil {
		t.Fatal("native duration allocation returned nil")
	}
	keep.Target48 = player
	remove.Target48 = monster
	srv.Spells.Dur.Add(keep)
	srv.Spells.Dur.Add(remove)

	spellDurationSelectiveCleanupExportCall4FE8A0(math.MinInt32)

	if srv.Spells.Dur.List != keep || keep.Prev != nil || keep.Next != nil {
		t.Fatalf("kept list = head %p prev %p next %p, want %p/nil/nil", srv.Spells.Dur.List, keep.Prev, keep.Next, keep)
	}
}

func TestSpellDurationSelectiveCleanupExport4FE8A0AcceptsNilAllocator(t *testing.T) {
	srv := useSpellDurationSelectiveCleanupLegacyServer4FE8A0(t)
	srv.Spells.Dur.List = &server.DurSpell{ID: 0x1234}
	spellDurationSelectiveCleanupExportCall4FE8A0(0)
	if srv.Spells.Dur.List != nil {
		t.Fatalf("duration list = %p, want nil", srv.Spells.Dur.List)
	}
}

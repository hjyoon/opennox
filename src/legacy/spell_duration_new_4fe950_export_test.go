package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type spellDurationNewLegacyServer4FE950 struct {
	Server
	srv *server.Server
}

func (s *spellDurationNewLegacyServer4FE950) S() *server.Server {
	return s.srv
}

func useSpellDurationNewLegacyServer4FE950(t *testing.T, active **server.Server) {
	t.Helper()
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationNewLegacyServer4FE950{srv: *active}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
}

func TestSpellDurationNewExport4FE950UsesLiveNativeServerAndPointer(t *testing.T) {
	firstServer := new(server.Server)
	secondServer := new(server.Server)
	for name, srv := range map[string]*server.Server{
		"first":  firstServer,
		"second": secondServer,
	} {
		if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
			t.Fatalf("%s allocator result = %d, want canonical 1", name, got)
		}
		t.Cleanup(srv.Spells.Dur.SpellFreeDurations4FE880)
	}
	active := firstServer
	useSpellDurationNewLegacyServer4FE950(t, &active)

	first := (*server.DurSpell)(spellDurationNewExportCall4FE950())
	second := (*server.DurSpell)(spellDurationNewExportCall4FE950())
	active = secondServer
	other := (*server.DurSpell)(spellDurationNewExportCall4FE950())
	if first == nil || second == nil || other == nil {
		t.Fatalf("CGo results = (%p, %p, %p), want three non-nil records", first, second, other)
	}
	if first.ID != 1 || second.ID != 2 || other.ID != 1 {
		t.Fatalf("CGo record IDs = (%d, %d, %d), want (1, 2, 1) across live servers", first.ID, second.ID, other.ID)
	}
	if got, want := *first, (server.DurSpell{ID: 1}); got != want {
		t.Fatalf("first CGo record = %#v, want allocator-zeroed record %#v", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, record := range map[string]*server.DurSpell{
			"first":  first,
			"second": second,
			"other":  other,
		} {
			if uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, record)
			}
		}
	}
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
	runtime.KeepAlive(other)
}

func TestSpellDurationNewExport4FE950FailureDoesNotConsumeID(t *testing.T) {
	srv := new(server.Server)
	active := srv
	useSpellDurationNewLegacyServer4FE950(t, &active)

	if got := spellDurationNewExportCall4FE950(); got != nil {
		t.Fatalf("uninitialized allocator result = %p, want nil", got)
	}
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(srv.Spells.Dur.SpellFreeDurations4FE880)
	record := (*server.DurSpell)(spellDurationNewExportCall4FE950())
	if record == nil || record.ID != 1 {
		var id uint16
		if record != nil {
			id = record.ID
		}
		t.Fatalf("post-failure record = %p ID %d, want non-nil ID 1", record, id)
	}
}

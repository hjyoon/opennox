package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellDurationUnlinkLegacyServer4FE900 struct {
	Server
	srv *server.Server
}

func (s *spellDurationUnlinkLegacyServer4FE900) S() *server.Server {
	return s.srv
}

func useSpellDurationUnlinkLegacyServer4FE900(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationUnlinkLegacyServer4FE900{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func TestSpellDurationUnlinkExport4FE900PreservesNativePointers(t *testing.T) {
	srv := useSpellDurationUnlinkLegacyServer4FE900(t)
	prev, freePrev := alloc.New(server.DurSpell{})
	record, freeRecord := alloc.New(server.DurSpell{})
	next, freeNext := alloc.New(server.DurSpell{})
	t.Cleanup(freePrev)
	t.Cleanup(freeRecord)
	t.Cleanup(freeNext)

	prev.Next = record
	record.Prev = prev
	record.Next = next
	next.Prev = record
	srv.Spells.Dur.List = prev

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"prev":   unsafe.Pointer(prev),
			"record": unsafe.Pointer(record),
			"next":   unsafe.Pointer(next),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	spellDurationUnlinkExportCall4FE900(unsafe.Pointer(record))

	if srv.Spells.Dur.List != prev || prev.Next != next || next.Prev != prev {
		t.Fatalf("links = head %p prev.Next %p next.Prev %p, want %p/%p/%p", srv.Spells.Dur.List, prev.Next, next.Prev, prev, next, prev)
	}
	if record.Prev != prev || record.Next != next {
		t.Fatalf("detached record links = %p/%p, want %p/%p unchanged", record.Prev, record.Next, prev, next)
	}
	runtime.KeepAlive(prev)
	runtime.KeepAlive(record)
	runtime.KeepAlive(next)
}

func TestSpellDurationUnlinkExport4FE900UpdatesHead(t *testing.T) {
	srv := useSpellDurationUnlinkLegacyServer4FE900(t)
	record, freeRecord := alloc.New(server.DurSpell{})
	next, freeNext := alloc.New(server.DurSpell{})
	t.Cleanup(freeRecord)
	t.Cleanup(freeNext)
	record.Next = next
	next.Prev = record
	srv.Spells.Dur.List = record

	spellDurationUnlinkExportCall4FE900(unsafe.Pointer(record))

	if srv.Spells.Dur.List != next || next.Prev != nil {
		t.Fatalf("head/next.Prev = %p/%p, want %p/nil", srv.Spells.Dur.List, next.Prev, next)
	}
	runtime.KeepAlive(record)
	runtime.KeepAlive(next)
}

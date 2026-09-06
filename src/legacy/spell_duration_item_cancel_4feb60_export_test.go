package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type itemCancelDurSpellsLegacyServer4FEB60 struct {
	Server
	srv *server.Server
}

func (s *itemCancelDurSpellsLegacyServer4FEB60) S() *server.Server {
	return s.srv
}

func useItemCancelDurSpellsLegacyServer4FEB60(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &itemCancelDurSpellsLegacyServer4FEB60{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func requireItemCancelDurSpellsLegacyPointers4FEB60(t *testing.T, values ...unsafe.Pointer) {
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

func TestItemCancelDurSpellsExport4FEB60PreservesNativePointers(t *testing.T) {
	srv := useItemCancelDurSpellsLegacyServer4FEB60(t)
	owner, freeOwner := alloc.New(server.Object{})
	item, freeItem := alloc.New(server.Object{})
	record43, free43 := alloc.New(server.DurSpell{})
	record59, free59 := alloc.New(server.DurSpell{})
	t.Cleanup(freeOwner)
	t.Cleanup(freeItem)
	t.Cleanup(free43)
	t.Cleanup(free59)

	item.ObjClass = object.Class(0x1000)
	item.ObjSubClass = object.SubClass(0x40000 | 0x4000000)
	record43.Spell = 43
	record43.Caster16 = owner
	record43.Flags88 = 0x12345620
	record43.Next = record59
	record59.Spell = 59
	record59.Caster16 = owner
	record59.Flags88 = 0x89abcdee
	srv.Spells.Dur.List = record43
	requireItemCancelDurSpellsLegacyPointers4FEB60(t,
		unsafe.Pointer(owner), unsafe.Pointer(item), unsafe.Pointer(record43), unsafe.Pointer(record59),
	)

	itemCancelDurSpellsExportCall4FEB60(owner, item)

	if record43.Flags88 != 0x12345621 || record59.Flags88 != 0x89abcdef {
		t.Fatalf("flags = %#x/%#x, want 0x12345621/0x89abcdef", record43.Flags88, record59.Flags88)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(record43)
	runtime.KeepAlive(record59)
}

func TestItemCancelDurSpellsExport4FEB60ClassGateAndNilOwner(t *testing.T) {
	srv := useItemCancelDurSpellsLegacyServer4FEB60(t)
	item, freeItem := alloc.New(server.Object{})
	record, freeRecord := alloc.New(server.DurSpell{})
	t.Cleanup(freeItem)
	t.Cleanup(freeRecord)

	item.ObjSubClass = object.SubClass(0x40000)
	record.Spell = 43
	record.Flags88 = 0x76543210
	srv.Spells.Dur.List = record
	requireItemCancelDurSpellsLegacyPointers4FEB60(t, unsafe.Pointer(item), unsafe.Pointer(record))

	itemCancelDurSpellsExportCall4FEB60(nil, item)
	if record.Flags88 != 0x76543210 {
		t.Fatalf("wrong-class flags = %#x, want 0x76543210", record.Flags88)
	}

	item.ObjClass = object.Class(0x1000)
	itemCancelDurSpellsExportCall4FEB60(nil, item)
	if record.Flags88 != 0x76543211 {
		t.Fatalf("nil-owner flags = %#x, want 0x76543211", record.Flags88)
	}
	runtime.KeepAlive(item)
	runtime.KeepAlive(record)
}

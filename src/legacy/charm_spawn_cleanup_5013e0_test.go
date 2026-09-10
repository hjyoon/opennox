package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestCharmQuestSpawnCleanup5013E0ResolvesNativePointersFromPE32Tokens(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	pool := alloc.NewClass("CharmSpawnCleanup", unsafe.Sizeof(charmSpawnLink5013E0{}), 3)
	t.Cleanup(pool.Free)
	poolToken := uint32(uintptr(pool.UPtr()))

	olderPointer := pool.NewObject()
	entryPointer := pool.NewObject()
	newerPointer := pool.NewObject()
	older := (*charmSpawnLink5013E0)(olderPointer)
	entry := (*charmSpawnLink5013E0)(entryPointer)
	newer := (*charmSpawnLink5013E0)(newerPointer)
	olderToken := uint32(uintptr(olderPointer))
	entryToken := uint32(uintptr(entryPointer))
	newerToken := uint32(uintptr(newerPointer))
	older.Newer = entryToken
	entry.Older = olderToken
	entry.Newer = newerToken
	newer.Older = entryToken

	generatorUpdate := new(server.MonsterGenUpdateData)
	generatorUpdate.ActiveCount = 3
	generator := &server.Object{UpdateData: unsafe.Pointer(generatorUpdate)}
	update := new(server.MonsterUpdateData)
	update.Field548 = generator
	update.Field549 = entryToken
	target := &server.Object{UpdateData: unsafe.Pointer(update)}
	var storedHead uint32
	charmQuestSpawnCleanupNative5013E0(target, poolToken, func(head uint32) {
		storedHead = head
	})

	if generatorUpdate.ActiveCount != 2 {
		t.Fatalf("active count = %d, want 2", generatorUpdate.ActiveCount)
	}
	if update.Field548 != nil || update.Field549 != 0 {
		t.Fatalf("spawn back-references = %p/%#x, want nil/zero", update.Field548, update.Field549)
	}
	if older.Newer != newerToken || newer.Older != olderToken {
		t.Fatalf("relinked tokens = %#x/%#x, want %#x/%#x",
			older.Newer, newer.Older, newerToken, olderToken)
	}
	if storedHead != 0 {
		t.Fatalf("middle unlink changed head to %#x", storedHead)
	}
	if got := pool.ActivePointerLegacy32(entryToken); got != nil {
		t.Fatalf("freed entry still active at %p", got)
	}

	newerUpdate := new(server.MonsterUpdateData)
	newerUpdate.Field549 = newerToken
	newerTarget := &server.Object{UpdateData: unsafe.Pointer(newerUpdate)}
	charmQuestSpawnCleanupNative5013E0(newerTarget, poolToken, func(head uint32) {
		storedHead = head
	})

	if older.Newer != 0 {
		t.Fatalf("tail link = %#x, want zero", older.Newer)
	}
	if storedHead != olderToken {
		t.Fatalf("new head = %#x, want %#x", storedHead, olderToken)
	}
	if newerUpdate.Field549 != 0 {
		t.Fatalf("newer spawn back-reference = %#x, want zero", newerUpdate.Field549)
	}
	if got := pool.ActivePointerLegacy32(newerToken); got != nil {
		t.Fatalf("freed newer entry still active at %p", got)
	}
}

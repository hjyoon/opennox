package legacy

/*
#include <stdint.h>

extern void* nox_alloc_spawn_2386216;
extern uint32_t dword_5d4594_2386212;
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// charmSpawnLink5013E0 deliberately retains the PE32 three-dword layout. The
// legacy spawn list stores low-dword identity tokens, not native Go pointers.
type charmSpawnLink5013E0 struct {
	Object uint32
	Older  uint32
	Newer  uint32
}

func charmQuestSpawnCleanupNative5013E0(
	target *server.Object,
	poolToken uint32,
	storeHead func(uint32),
) {
	if target == nil {
		return
	}
	update := (*server.MonsterUpdateData)(target.UpdateData)
	if generator := update.Field548; generator != nil {
		generatorUpdate := (*server.MonsterGenUpdateData)(generator.UpdateData)
		generatorUpdate.ActiveCount--
		update.Field548 = nil
	}

	entryToken := update.Field549
	if entryToken == 0 {
		return
	}
	pool := alloc.AsClassLegacy32(poolToken)
	if pool == nil {
		panic(fmt.Errorf("charm spawn pool token is not active: %#x", poolToken))
	}
	entryPointer := pool.ActivePointerLegacy32(entryToken)
	if entryPointer == nil {
		panic(fmt.Errorf("charm spawn entry token is not active: %#x", entryToken))
	}
	entry := (*charmSpawnLink5013E0)(entryPointer)
	if entry.Older != 0 {
		olderPointer := pool.ActivePointerLegacy32(entry.Older)
		if olderPointer == nil {
			panic(fmt.Errorf("charm older spawn token is not active: %#x", entry.Older))
		}
		(*charmSpawnLink5013E0)(olderPointer).Newer = entry.Newer
	}
	if entry.Newer != 0 {
		newerPointer := pool.ActivePointerLegacy32(entry.Newer)
		if newerPointer == nil {
			panic(fmt.Errorf("charm newer spawn token is not active: %#x", entry.Newer))
		}
		(*charmSpawnLink5013E0)(newerPointer).Older = entry.Older
	} else {
		storeHead(entry.Older)
	}
	pool.FreeObjectFirst(entryPointer)
	update.Field549 = 0
}

func charmQuestSpawnCleanupRuntime5013E0(target *server.Object) {
	charmQuestSpawnCleanupNative5013E0(
		target,
		uint32(uintptr(C.nox_alloc_spawn_2386216)),
		func(head uint32) {
			C.dword_5d4594_2386212 = C.uint32_t(head)
		},
	)
}

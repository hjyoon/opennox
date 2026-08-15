package legacy

/*
typedef struct nox_object_t nox_object_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Nox_xxx_respawnRemove_4EC6A0 restores GAME.EXE 004EC6A0 while retaining
// the native respawn record representation used by the surrounding legacy
// routines.
func Nox_xxx_respawnRemove_4EC6A0(obj *server.Object) {
	server.RespawnRemove4EC6A0(obj, server.RespawnRemoveHooks4EC6A0[
		*server.Object, *respawnRecord4EC5E0, unsafe.Pointer,
	]{
		LoadHead: respawnAddLoadHead4EC5E0,
		LoadObject: func(rec *respawnRecord4EC5E0) *server.Object {
			return rec.Object
		},
		LoadNext: func(rec *respawnRecord4EC5E0) *respawnRecord4EC5E0 {
			return rec.Next
		},
		LoadPrev: func(rec *respawnRecord4EC5E0) *respawnRecord4EC5E0 {
			return rec.Prev
		},
		StoreHead: respawnAddStoreHead4EC5E0,
		StoreNext: func(rec, next *respawnRecord4EC5E0) {
			rec.Next = next
		},
		StorePrev: func(rec, prev *respawnRecord4EC5E0) {
			rec.Prev = prev
		},
		LoadAllocator: respawnAddLoadAllocator4EC5E0,
		FreeFirst: func(allocator unsafe.Pointer, rec *respawnRecord4EC5E0) {
			alloc.AsClass(allocator).FreeObjectFirst(unsafe.Pointer(rec))
		},
	})
}

//export sub_4EC6A0
func sub_4EC6A0(obj *C.nox_object_t) {
	Nox_xxx_respawnRemove_4EC6A0(asObjectS((*nox_object_t)(obj)))
}

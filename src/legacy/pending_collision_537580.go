package legacy

/*
#include "defs.h"
#include "fist_collide_4eadf0.h"
#include "undead_killer_collide_4ebd40.h"
extern uint32_t dword_5d4594_2488604;
extern uint32_t dword_5d4594_2488608;
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var pendingCollisionQueue537580 server.PendingCollisionQueue537580
var pendingCollisionRuntime537610 = server.PendingCollisionRuntime537610{
	TypeByName:   func(name string) uint32 { return uint32(GetServer().S().Types.IndByID(name)) },
	Fist:         unsafe.Pointer(C.nox_xxx_collideFist_4EADF0),
	UndeadKiller: unsafe.Pointer(C.nox_xxx_collideUndeadKiller_4EBD40),
	PreparePath:  sub_50B500,
}

func pendingCollisionSync537580() {
	C.dword_5d4594_2488604 = 0
	C.dword_5d4594_2488608 = 0
	if pendingCollisionQueue537580.Head() != nil {
		C.dword_5d4594_2488604 = 1
	}
	if pendingCollisionQueue537580.Tail() != nil {
		C.dword_5d4594_2488608 = 1
	}
}
func pendingCollisionEnqueue537610(o *server.Object) byte {
	result := pendingCollisionQueue537580.Enqueue(o, pendingCollisionRuntime537610)
	pendingCollisionSync537580()
	return result
}

//export nox_pending_collision_contains_native_537580
func nox_pending_collision_contains_native_537580(o *nox_object_t) C.int {
	return C.int(pendingCollisionQueue537580.Contains(asObjectS(o)))
}

//export nox_pending_collision_empty_native_537590
func nox_pending_collision_empty_native_537590() C.int {
	if pendingCollisionQueue537580.Empty() {
		return 1
	}
	return 0
}

//export nox_pending_collision_remove_native_5375A0
func nox_pending_collision_remove_native_5375A0(o *nox_object_t) {
	pendingCollisionQueue537580.Remove(asObjectS(o))
	pendingCollisionSync537580()
}

//export nox_pending_collision_enqueue_native_537610
func nox_pending_collision_enqueue_native_537610(o *nox_object_t) C.char {
	return C.char(pendingCollisionEnqueue537610(asObjectS(o)))
}

//export nox_pending_collision_pop_native_537700
func nox_pending_collision_pop_native_537700() *nox_object_t {
	o := pendingCollisionQueue537580.Pop()
	pendingCollisionSync537580()
	return asObjectC(o)
}

//export nox_pending_collision_head_native_537740
func nox_pending_collision_head_native_537740() *nox_object_t {
	return asObjectC(pendingCollisionQueue537580.Head())
}

//export nox_pending_collision_next_native_537750
func nox_pending_collision_next_native_537750(o *nox_object_t) C.uintptr_t {
	if o == nil {
		return 0
	}
	if next, known := pendingCollisionQueue537580.Next(asObjectS(o)); known {
		return C.uintptr_t(uintptr(unsafe.Pointer(next)))
	}
	// Preserve an unowned object's raw PE32 link/sentinel as an integer. It is
	// never materialized as an invalid pointer on a Go stack.
	return C.uintptr_t(o.field_115)
}

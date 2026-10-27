package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
nox_object_t* sub_537700(void);
int sub_537590(void);
extern uint32_t dword_5d4594_2488604;
extern uint32_t dword_5d4594_2488608;
static inline uintptr_t nox_pending_fixture_next_word(nox_object_t* o) { return (uintptr_t)sub_537750(o); }
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var pendingProbeLookup537610 func(string) uint32
var pendingProbePrepare537610 func()

//export nox_pending_probe_lookup_537610
func nox_pending_probe_lookup_537610(name *C.char) C.uint32_t {
	return C.uint32_t(pendingProbeLookup537610(C.GoString(name)))
}

//export nox_pending_probe_prepare_537610
func nox_pending_probe_prepare_537610() { pendingProbePrepare537610() }

type pendingMachine537580 struct {
	enqueue  func(*server.Object) uint32
	contains func(*server.Object) uint32
	empty    func() uint32
	remove   func(*server.Object)
	pop      func() *server.Object
	head     func() *server.Object
	tail     func() uintptr
	next     func(*server.Object) uintptr
	link     func(*server.Object) uintptr
	compat   func() [2]uint32
	close    func()
}

func pendingProbeMachine537580(fist, undead unsafe.Pointer) pendingMachine537580 {
	oldQueue, oldRuntime := pendingCollisionQueue537580, pendingCollisionRuntime537610
	oldHead, oldTail := C.dword_5d4594_2488604, C.dword_5d4594_2488608
	pendingCollisionQueue537580 = server.PendingCollisionQueue537580{}
	pendingCollisionRuntime537610 = server.PendingCollisionRuntime537610{Fist: fist, UndeadKiller: undead,
		TypeByName:  func(name string) uint32 { return pendingProbeLookup537610(name) },
		PreparePath: func() { pendingProbePrepare537610() },
	}
	pendingCollisionSync537580()
	return pendingMachine537580{
		enqueue: func(o *server.Object) uint32 {
			return uint32(uint8(C.nox_xxx_unitHasCollideOrUpdateFn_537610(asObjectC(o))))
		},
		contains: func(o *server.Object) uint32 { return uint32(C.sub_537580(asObjectC(o))) },
		empty:    func() uint32 { return uint32(C.sub_537590()) },
		remove:   func(o *server.Object) { C.sub_5375A0(asObjectC(o)) },
		pop:      func() *server.Object { return asObjectS(C.sub_537700()) },
		head:     func() *server.Object { return asObjectS(C.sub_537740()) },
		tail:     func() uintptr { return uintptr(unsafe.Pointer(pendingCollisionQueue537580.Tail())) },
		next:     func(o *server.Object) uintptr { return uintptr(C.nox_pending_fixture_next_word(asObjectC(o))) },
		link: func(o *server.Object) uintptr {
			if next, known := pendingCollisionQueue537580.Next(o); known {
				return uintptr(unsafe.Pointer(next))
			}
			return uintptr(o.Field115)
		},
		compat: func() [2]uint32 { return [2]uint32{uint32(C.dword_5d4594_2488604), uint32(C.dword_5d4594_2488608)} },
		close: func() {
			pendingCollisionQueue537580, pendingCollisionRuntime537610 = oldQueue, oldRuntime
			C.dword_5d4594_2488604, C.dword_5d4594_2488608 = oldHead, oldTail
		},
	}
}

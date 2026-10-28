package legacy

/*
#include "defs.h"
#include "GAME4_3.h"
static uintptr_t mm53bda0_seen[3];
static int mm53bda0_collision(nox_object_t* missile, void* a, void* b) {
 mm53bda0_seen[0]=(uintptr_t)missile;
 mm53bda0_seen[1]=(uintptr_t)a;
 mm53bda0_seen[2]=(uintptr_t)b;
 return (int)0x89abcdefu;
}
static void* mm53bda0_collision_ptr(void) { return (void*)mm53bda0_collision; }
static uintptr_t mm53bda0_collision_seen(int n) { return mm53bda0_seen[n]; }
static int mm53bda0_probe(nox_object_t* missile) {
 return ((int(*)(nox_object_t*))nox_xxx_updateMagicMissile_53BDA0)(missile);
}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

func magicMissileCProbe53BDA0(missile *server.Object) int32 {
	return int32(C.mm53bda0_probe(asObjectC(missile)))
}
func magicMissileCollisionPointer53BDA0() unsafe.Pointer { return C.mm53bda0_collision_ptr() }
func magicMissileCollisionSeen53BDA0(n int) uintptr {
	return uintptr(C.mm53bda0_collision_seen(C.int(n)))
}

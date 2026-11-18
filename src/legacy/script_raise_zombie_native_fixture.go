package legacy

/*
#include "script_raise_zombie_516d00.h"
#include <stddef.h>
#include <string.h>

typedef uintptr_t (*script_raise_zombie_fn)(nox_object_t*);
_Static_assert(_Generic(&sub_516D00, script_raise_zombie_fn:1, default:0),
               "native mixed pointer/scalar return for sub_516D00");
static uintptr_t script_raise_zombie_call_C(int route, nox_object_t* unit) {
    script_raise_zombie_fn volatile fn = sub_516D00;
    return route ? fn(unit) : sub_516D00(unit);
}
static uintptr_t script_raise_zombie_data_C(nox_object_t* unit) {
    return (uintptr_t)unit->data_update;
}
static uint32_t script_raise_zombie_status_C(void* update, size_t offset) {
    uint32_t result;
    memcpy(&result, (unsigned char*)update + offset, sizeof(result));
    return result;
}
static void script_raise_zombie_layout_C(uintptr_t* result) {
    result[0] = sizeof(void*);
    result[1] = sizeof(nox_object_t);
    result[2] = offsetof(nox_object_t, obj_class);
    result[3] = offsetof(nox_object_t, obj_flags);
    result[4] = offsetof(nox_object_t, data_update);
    result[5] = sizeof(sub_516D00(NULL));
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func scriptRaiseZombieNativeCall(route int, unit *server.Object) uintptr {
	return uintptr(C.script_raise_zombie_call_C(C.int(route), asObjectC(unit)))
}

func scriptRaiseZombieNativeData(unit *server.Object) uintptr {
	return uintptr(C.script_raise_zombie_data_C(asObjectC(unit)))
}

func scriptRaiseZombieNativeStatus(update *server.MonsterUpdateData) uint32 {
	return uint32(C.script_raise_zombie_status_C(unsafe.Pointer(update), C.size_t(unsafe.Offsetof(update.StatusFlags))))
}

func scriptRaiseZombieNativeLayout() (result [6]uintptr) {
	C.script_raise_zombie_layout_C((*C.uintptr_t)(unsafe.Pointer(&result[0])))
	return
}

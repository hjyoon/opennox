package legacy

/*
#include "GAME3_3.h"

static int nox_object_damage_native_probe(
	nox_object_t* target, nox_object_t* source, nox_object_t* weapon,
	int32_t damage, int32_t damage_type) {
	(void)target;
	(void)source;
	(void)weapon;
	(void)damage;
	(void)damage_type;
	return -1;
}

static void* nox_object_damage_native_probe_ptr(void) {
	return nox_object_damage_native_probe;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func objectDamageNativeProbePtr() unsafe.Pointer {
	return C.nox_object_damage_native_probe_ptr()
}

func objectDamageDispatchCallNative(
	target, source, weapon *server.Object,
	damage int32,
	typ object.DamageType,
) bool {
	return C.nox_object_call_damage_native(
		asObjectC(target), asObjectC(source), asObjectC(weapon), C.int(damage), C.int(typ),
	) != 0
}

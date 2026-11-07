package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import "unsafe"

//export sub_4E24B0
func sub_4E24B0(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawStoneDamageNative4E24B0(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}

//export sub_4E24E0
func sub_4E24E0(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawMechGolemDamageNative4E24E0(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}

//export nox_xxx_damageFlammable_4E2520
func nox_xxx_damageFlammable_4E2520(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawFlammableDamageNative4E2520(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}

//export nox_xxx_damageBlackPowder_4E2560
func nox_xxx_damageBlackPowder_4E2560(target, source, weapon *C.nox_object_t, damage, typ C.int32_t) C.int32_t {
	return C.int32_t(rawBlackPowderDamageNative4E2560(unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(weapon), int32(damage), int32(typ)))
}

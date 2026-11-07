//go:build darwin && arm64

package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include "GAME3_3.h"
typedef int32_t (*raw_scalar_damage_native_fn)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t);
_Static_assert(sizeof(int32_t) == 4, "raw signed damage words");
_Static_assert(_Generic(&sub_4E24B0, raw_scalar_damage_native_fn: 1, default: 0), "Stone native ABI");
_Static_assert(_Generic(&sub_4E24E0, raw_scalar_damage_native_fn: 1, default: 0), "Mech native ABI");
_Static_assert(_Generic(&nox_xxx_damageFlammable_4E2520, raw_scalar_damage_native_fn: 1, default: 0), "Flammable native ABI");
_Static_assert(_Generic(&nox_xxx_damageBlackPowder_4E2560, raw_scalar_damage_native_fn: 1, default: 0), "BlackPowder native ABI");
static raw_scalar_damage_native_fn raw_scalar_damage_fixture_entry(uint32_t which) {
 switch (which) {
 case 0: return &sub_4E24B0;
 case 1: return &sub_4E24E0;
 case 2: return &nox_xxx_damageFlammable_4E2520;
 case 3: return &nox_xxx_damageBlackPowder_4E2560;
 default: abort();
 }
}
static int32_t raw_scalar_damage_fixture_call(uint32_t which, void* target, void* source, void* weapon, int32_t damage, int32_t typ) {
 return raw_scalar_damage_fixture_entry(which)((nox_object_t*)target, (nox_object_t*)source, (nox_object_t*)weapon, damage, typ);
}
static uintptr_t raw_scalar_damage_fixture_address(uint32_t which) { return (uintptr_t)raw_scalar_damage_fixture_entry(which); }
static uint32_t raw_scalar_damage_fixture_bad_u32(void* value) { return (uint32_t)(uintptr_t)value; }
static void* raw_scalar_damage_fixture_alloc(void) { return calloc(1, 32); }
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func rawScalarDamageFixtureCall4E24B0(which int, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	result := C.raw_scalar_damage_fixture_call(C.uint32_t(which), target, source, weapon, C.int32_t(damage), C.int32_t(typ))
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	return int32(result)
}
func rawScalarDamageFixtureAddress4E24B0(which int) uintptr {
	return uintptr(C.raw_scalar_damage_fixture_address(C.uint32_t(which)))
}
func rawScalarDamageFixtureBadU324E24B0(value unsafe.Pointer) uint32 {
	return uint32(C.raw_scalar_damage_fixture_bad_u32(value))
}
func rawScalarDamageFixtureAlloc4E24B0() unsafe.Pointer     { return C.raw_scalar_damage_fixture_alloc() }
func rawScalarDamageFixtureFree4E24B0(value unsafe.Pointer) { C.free(value) }

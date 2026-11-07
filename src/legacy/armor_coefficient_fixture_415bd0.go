//go:build darwin && arm64

package legacy

/*
#include "GAME1.h"
#include <stddef.h>
typedef double (*armor_coefficient_fn_415bd0)(nox_object_t*);
_Static_assert(_Generic(&sub_415BD0,armor_coefficient_fn_415bd0:1,default:0),"actual armor coefficient pointer ABI");
static double armor_coefficient_actual_C_415bd0(void* item) { armor_coefficient_fn_415bd0 fn=&sub_415BD0; return fn((nox_object_t*)item); }
static size_t armor_coefficient_class_offset_415bd0(void) { return offsetof(nox_object_t,obj_class); }
static size_t armor_coefficient_type_offset_415bd0(void) { return offsetof(nox_object_t,typ_ind); }
static double armor_coefficient_typed_C_conversion_415bd0(void* definition) { return (double)nox_modifier_getArmor(definition); }
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func armorCoefficientActualC415BD0(item unsafe.Pointer) float64 {
	result := float64(C.armor_coefficient_actual_C_415bd0(item))
	runtime.KeepAlive(item)
	return result
}

func armorCoefficientCOffsets415BD0() (class, typ uintptr) {
	return uintptr(C.armor_coefficient_class_offset_415bd0()), uintptr(C.armor_coefficient_type_offset_415bd0())
}

func armorCoefficientTypedCConversion415BD0(definition unsafe.Pointer) float64 {
	result := float64(C.armor_coefficient_typed_C_conversion_415bd0(definition))
	runtime.KeepAlive(definition)
	return result
}

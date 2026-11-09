package legacy

/*
#include "defs.h"
#include <stdint.h>
*/
import "C"
import "unsafe"

//export sub_415960
func sub_415960(query *C.wchar2_t) C.int {
	return C.int(titleToMask415960(titleWeapon415960, unsafe.Pointer(query)))
}

//export sub_4159F0
func sub_4159F0(mask C.int) *C.wchar2_t {
	return (*C.wchar2_t)(titleFromMask415960(titleWeapon415960, uint32(mask)))
}

//export sub_415DA0
func sub_415DA0(query *C.wchar2_t) C.int {
	return C.int(titleToMask415960(titleArmor415960, unsafe.Pointer(query)))
}

//export sub_415E80
func sub_415E80(mask C.int) *C.wchar2_t {
	return (*C.wchar2_t)(titleFromMask415960(titleArmor415960, uint32(mask)))
}

//export nox_native_loadModifyers_4158C0
func nox_native_loadModifyers_4158C0() C.uintptr_t {
	return C.uintptr_t(titleLoadTable415960(titleWeapon415960))
}

//export nox_native_loadLook_415D50
func nox_native_loadLook_415D50() C.uintptr_t {
	return C.uintptr_t(titleLoadTable415960(titleArmor415960))
}

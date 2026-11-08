package legacy

/*
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export sub_410550
func sub_410550(argument C.short) unsafe.Pointer {
	return secretWallLookupNative410550(uint32(argument))
}

//export sub_410730
func sub_410730() *C.nox_secret_wall_t {
	return (*C.nox_secret_wall_t)(secretWallClearNative410730())
}

//export sub_4107A0
func sub_4107A0(node unsafe.Pointer) *C.nox_secret_wall_t {
	return (*C.nox_secret_wall_t)(secretWallRemoveNative4107A0(node))
}

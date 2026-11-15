package legacy

/*
#include <stdint.h>
typedef const char nox_named_bitmask_char;
*/
import "C"
import "unsafe"

//export set_one_bitmask_flag_by_name_4239C0
func set_one_bitmask_flag_by_name_4239C0(name *C.char, bitmask *C.uint32_t, names **C.nox_named_bitmask_char) C.int {
	return C.int(namedBitmaskSetNative4239C0(unsafe.Pointer(name), (*uint32)(unsafe.Pointer(bitmask)), namedBitmaskNativeNames4239C0(unsafe.Pointer(names))))
}

//export set_bitmask_flags_from_plus_separated_names_423930
func set_bitmask_flags_from_plus_separated_names_423930(input *C.nox_named_bitmask_char, bitmask *C.uint32_t, names **C.nox_named_bitmask_char) {
	namedBitmaskParseNative423930(unsafe.Pointer(input), (*uint32)(unsafe.Pointer(bitmask)), namedBitmaskNativeNames4239C0(unsafe.Pointer(names)))
}

//export nox_native_parseMonsterFlags_423930
func nox_native_parseMonsterFlags_423930(input *C.nox_named_bitmask_char, bitmask *C.uint32_t) {
	namedBitmaskParseNative423930(unsafe.Pointer(input), (*uint32)(unsafe.Pointer(bitmask)), namedBitmaskMonsterNames4239C0)
}

package legacy

/*
#include <stdint.h>
typedef const char nox_soundset_field_char;
*/
import "C"
import "unsafe"

//export sub_4240F0
func sub_4240F0(data unsafe.Pointer, field *C.nox_soundset_field_char, value C.int) C.int {
	return C.int(soundSetFieldNative4240F0(data, unsafe.Pointer(field), uint32(value)))
}

//export nox_xxx_parseSoundSetBin_424170
func nox_xxx_parseSoundSetBin_424170(path *C.char) C.int {
	return C.int(soundSetParseNative424170(unsafe.Pointer(path)))
}

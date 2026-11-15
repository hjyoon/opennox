package legacy

/*
#include <stdint.h>
typedef const char nox_config_source_char;
*/
import "C"

import "unsafe"

//export sub_409B50
func sub_409B50(source *C.nox_config_source_char) C.uint {
	return C.uint(sub409B50Native(unsafe.Pointer(source)))
}

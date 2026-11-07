package legacy

/*
#include <stdint.h>
*/
import "C"

import "unsafe"

// The public const char* C ABI remains in the single C forwarder. This internal
// void* export avoids conflicting with cgo's generated non-const char* type.
//
//export sub_409B50_go
func sub_409B50_go(source unsafe.Pointer) C.uint32_t {
	return C.uint32_t(sub409B50Native(source))
}

package legacy

/*
#include <stdint.h>
#include "cgo_const.h"
*/
import "C"
import "unsafe"

//export sub_4134D0
func sub_4134D0() C.int { return C.int(sub4134D0Native()) }

//export sub_4137E0
func sub_4137E0() { sub4137E0Native() }

//export sub_413890
func sub_413890() *C.char { return (*C.char)(sub413890Native()) }

//export sub_4138E0
func sub_4138E0(a C.int) { sub4138E0Native(int32(a)) }

//export sub_413900
func sub_413900(a C.int) { sub413900Native(int32(a)) }

//export sub_413920
func sub_413920() C.int { return C.int(sub413920Native()) }

//export sub_4139B0
func sub_4139B0() C.int { return C.int(sub4139B0Native()) }

//export sub_413F60
func sub_413F60(a, b *C.nox_cgo_const_void) C.int {
	return C.int(sub413F60Native(unsafe.Pointer(a), unsafe.Pointer(b)))
}

package legacy

/*
#include "GAME1.h"
#include <stddef.h>
#include <stdint.h>
typedef int(*material7_fixture_fn_4133d0)(nox_object_t*);
_Static_assert(sizeof(int)==4,"Material7 public int return width");
_Static_assert(_Generic(&sub_4133D0,material7_fixture_fn_4133d0:1,default:0),"Material7 native object pointer ABI");
static int material7_fixture_call_4133d0(void*item){material7_fixture_fn_4133d0 fn=&sub_4133D0;return fn((nox_object_t*)item);}
static size_t material7_fixture_offset_4133d0(int field){switch(field){case 0:return offsetof(nox_object_t,init_data);case 1:return offsetof(nox_object_t,obj_class);default:return SIZE_MAX;}}
static void* material7_foreign_pointer_4133d0(uintptr_t bits){return(void*)bits;}
*/
import "C"

import (
	"runtime"
	"unsafe"
)

//export sub_4133D0
func sub_4133D0(item *C.nox_object_t) C.int {
	return C.int(rawMaterial7Native4133D0(unsafe.Pointer(item)))
}

func material7ForeignPointer4133D0(bits uintptr) unsafe.Pointer {
	return C.material7_foreign_pointer_4133d0(C.uintptr_t(bits))
}
func material7FixtureCall4133D0(item unsafe.Pointer) int32 {
	result := C.material7_fixture_call_4133d0(item)
	runtime.KeepAlive(item)
	return int32(result)
}
func material7FixtureCOffsets4133D0() (out [2]uintptr) {
	for i := range out {
		out[i] = uintptr(C.material7_fixture_offset_4133d0(C.int(i)))
	}
	return
}

package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
typedef const nox_object_t object_trace_const_object_t;
typedef const char object_trace_const_char_t;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_objectGetNetCode_4E4C00
func nox_xxx_objectGetNetCode_4E4C00(owner *C.object_trace_const_object_t) C.uint32_t {
	return C.uint32_t(objectTraceNetCode4E4C00(asObjectS((*nox_object_t)(unsafe.Pointer(owner)))))
}

//export nox_xxx_objectGetTypeInd_4E4C10
func nox_xxx_objectGetTypeInd_4E4C10(owner *C.object_trace_const_object_t) C.uint16_t {
	return C.uint16_t(objectTraceType4E4C10(asObjectS((*nox_object_t)(unsafe.Pointer(owner)))))
}

//export nox_object_getInitData_4E4C30
func nox_object_getInitData_4E4C30(owner *C.object_trace_const_object_t) unsafe.Pointer {
	return objectTraceInit4E4C30(asObjectS((*nox_object_t)(unsafe.Pointer(owner))))
}

//export nox_xxx_objectGetID_4E4C80
func nox_xxx_objectGetID_4E4C80(owner *C.object_trace_const_object_t) *C.object_trace_const_char_t {
	return (*C.object_trace_const_char_t)(objectTraceID4E4C80(asObjectS((*nox_object_t)(unsafe.Pointer(owner)))))
}

//export nox_xxx_inventoryGetNext_4E7990
func nox_xxx_inventoryGetNext_4E7990(owner *C.nox_object_t) *C.nox_object_t {
	next := objectInventoryNext4E7990(asObjectS((*nox_object_t)(owner)), func(obj *server.Object) *server.Object {
		return obj.InvNextItem
	})
	return (*C.nox_object_t)(asObjectC(next))
}

//export sub_537760
func sub_537760() *C.int2 {
	return (*C.int2)(unsafe.Pointer(objectTracePoint537760()))
}

//export sub_53B300
func sub_53B300(owner *C.nox_object_t) C.char {
	return C.char(objectTraceFlags53B300(asObjectS((*nox_object_t)(owner))))
}

package legacy

/*
#include "net_send_fx_all_523030.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

func netSendFxAllCliEntry523030(pos types.Pointf, data []byte) int32 {
	var ptr unsafe.Pointer
	if len(data) != 0 {
		ptr = unsafe.Pointer(unsafe.SliceData(data))
	}
	return int32(C.nox_xxx_netSendFxAllCli_523030(
		(*C.float2)(unsafe.Pointer(&pos)), ptr, C.int(len(data)),
	))
}

//export nox_xxx_netSendFxAllCliNative_523030
func nox_xxx_netSendFxAllCliNative_523030(pos *C.float2, data unsafe.Pointer, size C.int) C.int {
	if pos == nil || size < 0 || (data == nil && size != 0) {
		return 0
	}
	var packet []byte
	if size != 0 {
		packet = unsafe.Slice((*byte)(data), int(size))
	}
	GetServer().S().Nox_xxx_netSendFxAllCli_523030(
		types.Ptf(float32(pos.field_0), float32(pos.field_4)), packet,
	)
	return 0
}

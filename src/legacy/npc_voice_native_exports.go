package legacy

/*
#include "GAME1_1.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_setNPCVoiceSet_424320
func nox_xxx_setNPCVoiceSet_424320(obj *C.nox_object_t, data unsafe.Pointer) C.int {
	return C.int(setNPCVoiceSetNative424320((*server.Object)(unsafe.Pointer(obj)), data))
}

//export nox_getDefaultSoundSet_424350_native
func nox_getDefaultSoundSet_424350_native(name *C.char) unsafe.Pointer {
	if name == nil {
		return nil
	}
	return getDefaultSoundSet(GoString(name))
}

//export nox_xxx_readNPCVoiceSet_52AD10
func nox_xxx_readNPCVoiceSet_52AD10(obj *C.nox_object_t) C.size_t {
	return C.size_t(npcVoiceWireNative52AD10((*server.Object)(unsafe.Pointer(obj)), cryptfile.Global().ReadOnly(), func(buf []byte) uintptr {
		return uintptr(nox_xxx_fileReadWrite_426AC0_file3_fread_impl((*C.uchar)(unsafe.Pointer(unsafe.SliceData(buf))), C.size_t(len(buf)), nil, 0))
	}))
}

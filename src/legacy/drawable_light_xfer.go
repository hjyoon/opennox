package legacy

/*
#include <stdint.h>
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

//export nox_drawable_light_xfer_pack
func nox_drawable_light_xfer_pack(dr *C.nox_drawable, out *C.uint8_t) {
	if out == nil {
		return
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(out)), client.DrawableLightXferSize)
	clear(dst)
	if dr == nil {
		return
	}
	data := asDrawable(dr).LightXferData()
	copy(dst, data[:])
}

//export nox_drawable_light_xfer_unpack
func nox_drawable_light_xfer_unpack(dr *C.nox_drawable, in *C.uint8_t) {
	if dr == nil || in == nil {
		return
	}
	data := (*[client.DrawableLightXferSize]byte)(unsafe.Pointer(in))
	asDrawable(dr).ApplyLightXferData(data)
}

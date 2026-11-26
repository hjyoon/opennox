package legacy

/*
#include "GAME2_2.h"
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

//export nox_drawable_change_light_color_native
func nox_drawable_change_light_color_native(dr *nox_drawable, r, g, b C.int) *nox_drawable {
	asDrawable(dr).SetLightColorRGB(noxrender.RGB{R: int(r), G: int(g), B: int(b)})
	return dr
}

//export nox_xxx_spriteChangeLightColor_484BE0
func nox_xxx_spriteChangeLightColor_484BE0(light *C.uint32_t, r, g, b C.int) *C.uint32_t {
	data := unsafe.Slice((*byte)(unsafe.Pointer(light)), 28)
	binary.LittleEndian.PutUint32(data[16:20], uint32(r))
	binary.LittleEndian.PutUint32(data[:4], 2)
	binary.LittleEndian.PutUint32(data[20:24], uint32(g))
	binary.LittleEndian.PutUint32(data[24:28], uint32(b))
	return light
}

//export nox_drawable_change_light_direction_native
func nox_drawable_change_light_direction_native(dr *nox_drawable, direction C.int) C.longlong {
	return C.longlong(asDrawable(dr).SetLightDirection(int32(direction)))
}

//export nox_drawable_change_light_size_native
func nox_drawable_change_light_size_native(dr *nox_drawable, size C.int) C.longlong {
	return C.longlong(asDrawable(dr).SetLightPenumbra(int32(size)))
}

//export nox_drawable_change_light_intensity_native
func nox_drawable_change_light_intensity_native(dr *nox_drawable, intensity C.float) C.int {
	drawable := asDrawable(dr)
	if drawable == nil {
		return 0
	}
	drawable.SetLightIntensity(float32(intensity))
	return C.int(drawable.LightIntensityRad)
}

//export sub_484CE0
func sub_484CE0(light unsafe.Pointer, intensity C.float) C.int {
	value := float32(intensity)
	if value > 63 {
		value = 63
	}
	data := unsafe.Slice((*byte)(light), 12)
	binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(value))
	radius := client.LightRadius(value)
	binary.LittleEndian.PutUint32(data[8:12], uint32(radius))
	return C.int(radius)
}

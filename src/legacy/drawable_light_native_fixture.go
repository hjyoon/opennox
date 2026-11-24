package legacy

/*
#include "GAME2_2.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func drawableLightColorCABI(dr *client.Drawable, color noxrender.RGB) *client.Drawable {
	p := (*nox_drawable)(unsafe.Pointer(dr))
	return asDrawable(C.nox_drawable_change_light_color_native(p, C.int(color.R), C.int(color.G), C.int(color.B)))
}

func drawableLightAngleCABI(dr *client.Drawable, value int32, direction bool) int64 {
	p := (*nox_drawable)(unsafe.Pointer(dr))
	if direction {
		return int64(C.nox_drawable_change_light_direction_native(p, C.int(value)))
	}
	return int64(C.nox_drawable_change_light_size_native(p, C.int(value)))
}

func drawableLightIntensityCABI(dr *client.Drawable, packed []byte, value float32) (int, int) {
	p := (*nox_drawable)(unsafe.Pointer(dr))
	return int(C.nox_drawable_change_light_intensity_native(p, C.float(value))),
		int(C.sub_484CE0(unsafe.Pointer(&packed[0]), C.float(value)))
}

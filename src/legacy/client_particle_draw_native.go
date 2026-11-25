package legacy

/*
#include "client__draw__harpoondraw.h"
#include "client__draw__lvupdraw.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export nox_thing_harpoon_draw
func nox_thing_harpoon_draw(view *C.int, drawable *C.nox_drawable) C.int {
	return nox_thing_slave_draw((*C.uint32_t)(unsafe.Pointer(view)), drawable)
}

func harpoonRopeEndpoint(code uint16) *nox_drawable {
	if nox_xxx_netTestHighBit_578B70(C.uint(code)) != 0 {
		return nox_xxx_netSpriteByCodeStatic_45A720(int32(code & 0x7fff))
	}
	return nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(code))
}

//export nox_thing_harpoon_rope_draw
func nox_thing_harpoon_rope_draw(view *C.int, drawable *C.nox_drawable) C.int {
	viewport := (*C.nox_draw_viewport_t)(unsafe.Pointer(view))
	ray := drawableEffectBytes(asDrawable(drawable))
	sourceCode := binary.LittleEndian.Uint32(ray[5:9])
	targetCode := binary.LittleEndian.Uint32(ray[9:13])
	x := int32(viewport.x1) - int32(viewport.field_4)
	y := int32(viewport.y1) - int32(viewport.field_5)
	var source, target C.int2
	if ray[0] == 0 {
		source.field_0 = C.int(x + int32(uint16(sourceCode)))
		source.field_4 = C.int(y + int32(sourceCode>>16) - 20)
		target.field_0 = C.int(x + int32(uint16(targetCode)))
		target.field_4 = C.int(y + int32(targetCode>>16) - 20)
	} else {
		from := harpoonRopeEndpoint(uint16(sourceCode))
		to := harpoonRopeEndpoint(uint16(targetCode))
		if from == nil || to == nil {
			return 1
		}
		offset := 8 * uintptr(from.field_74_2)
		source.field_0 = C.int(x + int32(from.pos.x) + int32(memmap.Uint32(0x587000, 175864+offset)))
		source.field_4 = C.int(y + int32(from.pos.y) + int32(memmap.Uint32(0x587000, 175868+offset)))
		target.field_0 = C.int(x + int32(to.pos.x))
		target.field_4 = C.int(y + int32(to.pos.y) - int32(int16(to.field_26_1)) - int32(int16(to.z)) - 8)
	}
	*memmap.PtrUint32(0x5D4594, 1312492) = game3EffectColor(144, 104, 64)
	*memmap.PtrUint32(0x5D4594, 1312496) = game3EffectColor(24, 16, 0)
	sub_4B63B0(&source, &target)
	return 1
}

//export nox_thing_falling_sparks_draw_4B7740
func nox_thing_falling_sparks_draw_4B7740(typeID C.int, view *C.nox_draw_viewport_t, drawable *C.nox_drawable) *C.uint32_t {
	file := CString(`C:\NoxPost\src\client\Draw\LvUpDraw.c`)
	var last *nox_drawable
	for i := 0; i < 2; i++ {
		x := int32(drawable.pos.x) + nox_common_randomIntMinMax_415FF0(-15, 15, file, 35)
		y := int32(drawable.pos.y) + nox_common_randomIntMinMax_415FF0(-15, 15, file, 36)
		z := uint16(y - int32(view.field_5))
		v := nox_common_randomIntMinMax_415FF0(8, 12, file, 40)
		spawn := C.int2{field_0: C.int(x), field_4: C.int(y)}
		origin := C.int2{field_0: C.int(drawable.pos.x), field_4: C.int(drawable.pos.y)}
		last = sub_499950(typeID, &spawn, &origin, C.ushort(z), C.char(-v))
	}
	return (*C.uint32_t)(unsafe.Pointer(last))
}

package legacy

/*
#include "GAME1_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__video__draw_common.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const sentryRaysBase4C5020 = 1321540

func sentryRaySlot4C5020(index int) *[4]uint16 {
	return memmap.PtrT[[4]uint16](0x5D4594, uintptr(sentryRaysBase4C5020+8*index))
}

//export nox_xxx_colorInit_4C4FD0
func nox_xxx_colorInit_4C4FD0() C.int {
	*memmap.PtrUint32(0x5D4594, 1321536) = uint32(C.nox_color_rgb_4344A0(255, 200, 255))
	*memmap.PtrUint32(0x5D4594, 1321796) = uint32(C.nox_color_rgb_4344A0(255, 0, 255))
	color := C.nox_color_rgb_4344A0(100, 40, 100)
	*memmap.PtrUint32(0x5D4594, 1321532) = uint32(color)
	return C.int(color)
}

func addSentryRayNative4C5020(ray [4]uint16) int32 {
	count := int32(legacyAddressableWords.dword_5d4594_1321800)
	if count < 32 {
		*sentryRaySlot4C5020(int(count)) = ray
		count++
		legacyAddressableWords.dword_5d4594_1321800 = C.uint32_t(count)
	}
	return count
}

//export nox_client_addSentryRay_4C5020
func nox_client_addSentryRay_4C5020(x1, y1, x2, y2 C.uint16_t) C.int {
	return C.int(addSentryRayNative4C5020([4]uint16{uint16(x1), uint16(y1), uint16(x2), uint16(y2)}))
}

//export nox_client_sentryRayCount_4C5020
func nox_client_sentryRayCount_4C5020() C.int {
	return C.int(legacyAddressableWords.dword_5d4594_1321800)
}

//export nox_client_sentryRayAt_4C5020
func nox_client_sentryRayAt_4C5020(index C.int, ray *C.uint16_t) C.int {
	if ray == nil || index < 0 || index >= nox_client_sentryRayCount_4C5020() {
		return 0
	}
	*(*[4]uint16)(unsafe.Pointer(ray)) = *sentryRaySlot4C5020(int(index))
	return 1
}

//export sub_4C5020
func sub_4C5020(data *C.nox_sentry_const_u8_t) C.int {
	raw := unsafe.Slice((*byte)(unsafe.Pointer(data)), 9)
	var ray [4]uint16
	for i := range ray {
		ray[i] = binary.LittleEndian.Uint16(raw[1+i*2:])
	}
	return C.int(addSentryRayNative4C5020(ray))
}

func sentryRayScreenPosition4C5060(vp *C.nox_draw_viewport_t, ray [4]uint16) (from, to C.int2) {
	x, y := C.int(vp.x1)-C.int(vp.field_4), C.int(vp.y1)-C.int(vp.field_5)
	from.field_0, from.field_4 = x+C.int(ray[0]), y+C.int(ray[1])
	to.field_0, to.field_4 = x+C.int(ray[2]), y+C.int(ray[3])
	return from, to
}

//export nox_client_sentryRayScreenPosition_4C5060
func nox_client_sentryRayScreenPosition_4C5060(vp *C.nox_sentry_const_viewport_t, ray *C.nox_sentry_const_u16_t, from, to *C.int2) {
	*from, *to = sentryRayScreenPosition4C5060((*C.nox_draw_viewport_t)(vp), *(*[4]uint16)(unsafe.Pointer(ray)))
}

//export sub_4C5060
func sub_4C5060(vp *C.nox_draw_viewport_t) C.int {
	if *memmap.PtrPtr(0x852978, 8) == nil || nox_client_sentryRayCount_4C5020() == 0 {
		return 0
	}
	for index := 0; ; index++ {
		from, to := sentryRayScreenPosition4C5060(vp, *sentryRaySlot4C5020(index))
		x, y := from.field_0, from.field_4
		visible := C.sub_4992B0(x, y)
		if x <= 0 || x >= C.int(vp.width)-1 || y <= 0 || y >= C.int(vp.height)-1 {
			visible = 0
		}
		crossings := sub_498C20((*C.nox_point)(unsafe.Pointer(&from)), (*C.nox_point)(unsafe.Pointer(&to)), 1)
		for i := int32(0); i < crossings; i++ {
			point := sub_499290(i)
			next := C.int2{field_0: point.x, field_4: point.y}
			if visible != 0 {
				sub_4C51D0(&from, &next)
			}
			from = next
			visible = 1 - visible
		}
		if visible != 0 {
			sub_4C51D0(&from, &to)
		}
		if next := C.int(index + 1); next >= nox_client_sentryRayCount_4C5020() {
			return next
		}
	}
}

func sentryAbs4C51D0(value C.int) C.int {
	if value < 0 {
		return -value
	}
	return value
}

//export sub_4C51D0
func sub_4C51D0(from, to *C.int2) C.int {
	x, y := from.field_0, from.field_4
	dx, dy := to.field_0-x, to.field_4-y
	C.nox_client_drawSetColor_434460(C.int(memmap.Uint32(0x5D4594, 1321532)))
	C.nox_client_drawEnableAlpha_434560(1)
	C.nox_client_drawAddPoint_49F500(x, y)
	C.nox_xxx_rasterPointRel_49F570(dx, dy)
	C.nox_client_drawLineFromPoints_49E4B0()
	C.nox_client_drawEnableAlpha_434560(0)

	y -= 22
	C.nox_client_drawSetColor_434460(C.int(memmap.Uint32(0x5D4594, 1321536)))
	C.nox_client_drawAddPoint_49F500(x, y)
	C.nox_xxx_rasterPointRel_49F570(dx, dy)
	C.nox_client_drawLineFromPoints_49E4B0()

	C.nox_client_drawSetColor_434460(C.int(memmap.Uint32(0x5D4594, 1321796)))
	offsetX, offsetY := C.int(1), C.int(0)
	if sentryAbs4C51D0(dx) > sentryAbs4C51D0(dy) {
		offsetX, offsetY = 0, 1
	}
	C.nox_client_drawAddPoint_49F500(x-offsetX, y-offsetY)
	C.nox_client_drawAddPoint_49F500(x+dx-offsetX, y+dy-offsetY)
	C.nox_client_drawLineFromPoints_49E4B0()
	C.nox_client_drawAddPoint_49F500(x+offsetX, y+offsetY)
	C.nox_client_drawAddPoint_49F500(x+dx+offsetX, y+dy+offsetY)
	return C.nox_client_drawLineFromPoints_49E4B0()
}

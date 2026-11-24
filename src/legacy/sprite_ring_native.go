package legacy

/*
#include "GAME3_1.h"
*/
import "C"

import "github.com/opennox/opennox/v1/common/memmap"

func spriteRingVertex(radius int32, index uintptr) (x, y int32) {
	offset := 192088 + index*128
	x = (radius * memmap.Int32(0x587000, offset)) >> 4
	y = (radius * memmap.Int32(0x587000, offset+4)) >> 4
	return x, y
}

//export nox_xxx_spriteDrawCircleMB_4C32A0
func nox_xxx_spriteDrawCircleMB_4C32A0(x, y, radius, color C.int) C.int {
	cx, cy, r := int32(x), int32(y), int32(radius)
	prevX, prevY := spriteRingVertex(r, 0)
	nox_client_drawSetColor_434460(int32(color))
	nox_client_drawEnableAlpha_434560(1)
	for i := uintptr(1); i < 16; i++ {
		x, y := spriteRingVertex(r, i)
		nox_client_drawAddPoint_49F500(cx+prevX, cy+prevY)
		nox_client_drawAddPoint_49F500(cx+x, cy+y)
		nox_client_drawLineFromPoints_49E4B0()
		prevX, prevY = x, y
	}
	firstX, firstY := spriteRingVertex(r, 0)
	nox_client_drawAddPoint_49F500(cx+firstX, cy+firstY)
	nox_client_drawAddPoint_49F500(cx+prevX, cy+prevY)
	nox_client_drawLineFromPoints_49E4B0()
	nox_client_drawEnableAlpha_434560(0)
	return 1
}

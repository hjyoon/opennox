package legacy

/*
#include "GAME3.h"
extern uint32_t dword_5d4594_1313532;
extern uint32_t dword_5d4594_1313536;
extern uint32_t dword_5d4594_1313540;
extern uint32_t dword_5d4594_1313564;
*/
import "C"

import (
	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/opennox/v1/common/memmap"
)

func game3EffectColor(r, g, b int) uint32 {
	// The genuine RGB5551 service consumes the low byte of each component.
	// Undefined upper bytes of old LOBYTE temporaries never affect its result.
	return noxcolor.RGB5551Color(byte(r), byte(g), byte(b)).Color32()
}

//export sub_4B5CD0
func sub_4B5CD0() int32 {
	var result uint32
	for i := 0; i < 64; i++ {
		v := i * 255 / 63
		*memmap.PtrUint32(0x5D4594, 1312500+uintptr(4*i)) = game3EffectColor(v/3, 3*v/5, v)
	}
	for i := 0; i < 32; i++ {
		v := i * 255 / 32
		*memmap.PtrUint32(0x5D4594, 1312756+uintptr(4*i)) = game3EffectColor(v, v, 0)
		*memmap.PtrUint32(0x5D4594, 1312884+uintptr(4*i)) = game3EffectColor(255, -1-v, 0)
	}
	for i := 0; i < 64; i++ {
		v := i * 255 / 63
		*memmap.PtrUint32(0x5D4594, 1313012+uintptr(4*i)) = game3EffectColor(v, v, v)
	}
	for i := 0; i < 64; i++ {
		result = game3EffectColor(i*255/63, 50, 50)
		*memmap.PtrUint32(0x5D4594, 1313268+uintptr(4*i)) = result
	}
	return int32(result)
}

//export sub_4B64C0
func sub_4B64C0() int32 {
	colors := [...][3]int{
		{255, 255, 0}, {255, 100, 0}, {255, 255, 0}, {0, 0, 255},
		{0, 200, 255}, {0, 200, 200}, {50, 255, 255}, {255, 0, 255},
		{255, 200, 255}, {255, 200, 0}, {255, 255, 100}, {100, 255, 50},
		{150, 255, 150}, {255, 255, 0}, {0, 220, 0}, {150, 255, 150},
		{200, 200, 200}, {255, 255, 255},
	}
	for i, rgb := range colors {
		v := game3EffectColor(rgb[0], rgb[1], rgb[2])
		switch i {
		case 2:
			C.dword_5d4594_1313532 = C.uint32_t(v)
		case 3:
			C.dword_5d4594_1313536 = C.uint32_t(v)
		case 4:
			C.dword_5d4594_1313540 = C.uint32_t(v)
		case 10:
			C.dword_5d4594_1313564 = C.uint32_t(v)
		default:
			*memmap.PtrUint32(0x5D4594, 1313524+uintptr(4*i)) = v
		}
	}
	r, g, b := 255, 255, 255
	for i := 0; i < 16; i++ {
		switch {
		case i <= 3:
			b = (765 - 255*i) / 3
		case i <= 7:
			g = (600-200*i)/4 - 1
		default:
			r = (1085-155*i)/9 - 1
		}
		*memmap.PtrUint32(0x5D4594, 1313656-uintptr(4*i)) = game3EffectColor(r, g, b)
	}
	return sub_4B5CD0()
}

type game3LinePoint struct{ x, y int32 }

// game3HarpoonLines preserves the PE32 geometry and the three draw-call order.
// Keeping the scalar geometry separate also allows bounded original controls.
func game3HarpoonLines(a, b game3LinePoint, draw func(game3LinePoint, game3LinePoint, bool) int32) int32 {
	draw(a, b, false)
	dx, dy := b.x-a.x, b.y-a.y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx <= dy {
		draw(game3LinePoint{a.x - 1, a.y}, game3LinePoint{b.x - 1, b.y}, true)
		return draw(game3LinePoint{a.x + 1, a.y}, game3LinePoint{b.x + 1, b.y}, true)
	}
	draw(game3LinePoint{a.x, a.y - 1}, game3LinePoint{b.x, b.y - 1}, true)
	return draw(game3LinePoint{a.x, a.y + 1}, game3LinePoint{b.x, b.y + 1}, true)
}

//export sub_4B63B0
func sub_4B63B0(a, b *C.int2) int32 {
	colorSet, lastEdge := false, false
	return game3HarpoonLines(
		game3LinePoint{int32(a.field_0), int32(a.field_4)},
		game3LinePoint{int32(b.field_0), int32(b.field_4)},
		func(start, end game3LinePoint, edge bool) int32 {
			if !colorSet || edge != lastEdge {
				off := uintptr(1312492)
				if edge {
					off = 1312496
				}
				nox_client_drawSetColor_434460(int32(*memmap.PtrUint32(0x5D4594, off)))
				colorSet, lastEdge = true, edge
			}
			nox_client_drawAddPoint_49F500(start.x, start.y)
			nox_client_drawAddPoint_49F500(end.x, end.y)
			return nox_client_drawLineFromPoints_49E4B0()
		})
}

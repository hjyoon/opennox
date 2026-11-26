package legacy

/*
#include "GAME1_2.h"
#include "client__draw__partscrn.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

var screenParticleColors = [...][2]uintptr{
	{806016, 806036},
	{806028, 806004},
	{806032, 806040},
	{806020, 806012},
	{806008, 806024},
}

//export nox_client_newScreenParticle_431540
func nox_client_newScreenParticle_431540(color, x, y, vx, vy, accel C.int, radius, interval, mode, flags C.char) *C.nox_screenParticle {
	pool := (*C.nox_alloc_class)(legacyPointers.screenParticleAllocator)
	if pool == nil || color < 0 || int(color) >= len(screenParticleColors) {
		return nil
	}
	p := (*C.nox_screenParticle)(nox_alloc_class_new_obj_zero(pool))
	if p == nil {
		p = legacyPointers.screenParticleTail
		if p == nil {
			return nil
		}
		sub_4316C0(p)
	}
	palette := screenParticleColors[color]
	p.field_4 = C.uint32_t(color)
	p.field_8 = C.uint32_t(*g2ServerWord(palette[0]))
	p.field_12 = C.uint32_t(*g2ServerWord(palette[1]))
	p.field_16, p.field_20 = C.uint32_t(vx)<<16, C.uint32_t(vy)<<16
	p.field_24, p.field_28 = C.uint32_t(x)<<16, C.uint32_t(y)<<16
	p.field_32 = p.field_32&^0xff | C.uint32_t(byte(flags))
	p.field_36 = C.uint32_t(accel) << 16
	p.field_40 = [4]C.uint8_t{C.uint8_t(radius), C.uint8_t(interval), C.uint8_t(mode), C.uint8_t(interval)}
	p.draw_fnc = (*[0]byte)(C.nox_client_screenParticleDraw_489700)
	nox_client_addScreenParticle_431680(p)
	if p.field_36 == 0 && p.field_40[1] == 0 {
		p.field_40[1], p.field_40[2], p.field_40[3] = 3, 2, 3
	}
	return p
}

//export nox_client_addScreenParticle_431680
func nox_client_addScreenParticle_431680(p *C.nox_screenParticle) {
	head := legacyPointers.screenParticleHead
	p.field_44, p.field_48 = head, nil
	if head != nil {
		head.field_48 = p
	} else {
		legacyPointers.screenParticleTail = p
	}
	legacyPointers.screenParticleHead = p
}

//export sub_4316C0
func sub_4316C0(p *C.nox_screenParticle) {
	next, prev := p.field_44, p.field_48
	if p == legacyPointers.screenParticleTail {
		legacyPointers.screenParticleTail = prev
	}
	if next != nil {
		next.field_48 = prev
	}
	if prev != nil {
		prev.field_44 = next
	} else {
		legacyPointers.screenParticleHead = next
	}
}

//export sub_431700
func sub_431700(p *C.nox_screenParticle) {
	sub_4316C0(p)
	nox_alloc_class_free_obj_first((*C.nox_alloc_class)(legacyPointers.screenParticleAllocator), unsafe.Pointer(p))
}

//export nox_client_screenParticlesDraw_431720
func nox_client_screenParticlesDraw_431720(view *nox_draw_viewport_t) {
	if view == nil {
		return
	}
	vp := asViewport(view)
	C.sub_430B50(C.int(vp.Screen.Min.X), C.int(vp.Screen.Min.Y), C.int(vp.Screen.Max.X), C.int(vp.Screen.Max.Y))
	for p := legacyPointers.screenParticleHead; p != nil; {
		set_dword_5d4594_3799468(1)
		// The draw callback can release or recycle the current particle.
		next := p.field_44
		ccall.CallIntPtr2(unsafe.Pointer(p.draw_fnc), unsafe.Pointer(view), unsafe.Pointer(p))
		p = next
	}
}

const screenParticleSource = `C:\NoxPost\src\client\Draw\PartScrn.c`

func screenParticleRandom(min, max, line int32) int32 {
	return nox_common_randomIntMinMax_415FF0(min, max, internCStr(screenParticleSource), line)
}

//export nox_client_screenParticleDraw_489700
func nox_client_screenParticleDraw_489700(view unsafe.Pointer, p *C.nox_screenParticle) C.int {
	vp := asViewportP(view)
	x, y := int32(p.field_24>>16), int32(p.field_28>>16)
	if x <= 0 || y <= 0 || int(x) >= vp.Size.X || int(y) >= vp.Size.Y {
		sub_431700(p)
		return 0
	}
	pos := C.int2{field_0: C.int(x), field_4: C.int(y)}
	sub_4B6720(&pos, C.int(p.field_8), C.int(p.field_40[0]), C.char(p.field_40[0]))
	nox_client_drawSetColor_434460(int32(p.field_12))
	nox_xxx_drawPointMB_499B70(x, y, int32(p.field_40[0]>>1))

	p.field_20 += p.field_36
	if p.field_40[1] != 0 {
		p.field_40[1]--
		if p.field_40[1] == 0 {
			if p.field_40[2] == 1 {
				p.field_40[0]++
				if p.field_40[0] >= 4 {
					p.field_40[2] = 2
				}
			} else {
				p.field_40[0]--
				if p.field_40[0] == 0 {
					sub_431700(p)
					return 0
				}
			}
			p.field_40[1] = p.field_40[3]
		}
	}

	if byte(p.field_32) == 1 && screenParticleRandom(0, 10, 105) >= 8 {
		interval := C.char(screenParticleRandom(3, 5, 115))
		radius := C.char(screenParticleRandom(2, 3, 114))
		dx := C.int(screenParticleRandom(-2, 2, 109))
		nox_client_newScreenParticle_431540(C.int(p.field_4), pos.field_0+dx, pos.field_4, 0, 0, 1, radius, interval, 2, 2)
	}
	p.field_24 += p.field_16
	p.field_28 += p.field_20
	return 1
}

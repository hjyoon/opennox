package legacy

/*
#include "client__draw__fx.h"
#include "client__draw__lightning.h"
*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
)

const lightningDrawSource = `C:\NoxPost\src\Client\Draw\lightning.c`

func lightningPackPoint(p C.int2) C.uint32_t {
	return C.uint32_t(uint16(p.field_0)) | C.uint32_t(uint16(p.field_4))<<16
}

func lightningUnpackPoint(p C.uint32_t) C.int2 {
	return C.int2{field_0: C.int(int16(p)), field_4: C.int(int16(p >> 16))}
}

func lightningSpan(from, to C.int2) (dx, dy int32) {
	dx = int32(max(from.field_0, to.field_0) - min(from.field_0, to.field_0))
	dy = int32(max(from.field_4, to.field_4) - min(from.field_4, to.field_4))
	return dx, dy
}

func drawLightningSegment(from, to C.int2) {
	color := int32(legacyWords.dword_5d4594_1316472)
	thick := memmap.Uint32(0x5D4594, 1316508) != 0
	halo, intensity, extent := color, int32(3), int32(32)
	if thick {
		halo = memmap.Int32(0x5D4594, 1316440)
		intensity = 12
		extent = int32(memmap.Uint8(0x5D4594, 1316420)) + 48
	}
	nox_draw_set54RGB32_434040(halo)
	sub_434080(intensity)
	addDrawLinePoints(from, to)
	sub_49E4F0(extent)
	nox_client_drawSetColor_434460(color)
	addDrawLinePoints(from, to)
	nox_client_drawLineFromPoints_49E4B0()
	if !thick {
		return
	}

	dx, dy := lightningSpan(from, to)
	for _, side := range [...]C.int{1, -1} {
		for i := C.int(1); i < C.int(memmap.Uint8(0x5D4594, 1316420)>>1); i++ {
			nox_client_drawResetPoints_49F5A0()
			shift := C.int2{field_0: 0, field_4: side * i}
			if dx <= dy {
				shift.field_0, shift.field_4 = shift.field_4, 0
			}
			addDrawLinePoints(
				C.int2{field_0: to.field_0 + shift.field_0, field_4: to.field_4 + shift.field_4},
				C.int2{field_0: from.field_0 + shift.field_0, field_4: from.field_4 + shift.field_4},
			)
			nox_client_drawLineFromPoints_49E4B0()
		}
	}
}

//export nox_xxx_drawLightningStep_4BB070
func nox_xxx_drawLightningStep_4BB070(start, end C.uint32_t) C.int {
	legacyWords.dword_5d4594_1316492++
	depth := int32(legacyWords.dword_5d4594_1316492)
	from, to := lightningUnpackPoint(start), lightningUnpackPoint(end)
	if legacyWords.dword_5d4594_1316492 == legacyWords.dword_5d4594_1316448 {
		drawLightningSegment(from, to)
	} else {
		num, den := int32(1), int32(1)
		decay := int32(memmap.Uint32(0x587000, 178212) + 9)
		for i := int32(1); i < depth; i++ {
			num *= 10
			den *= decay
		}
		jitter := int32(legacyWords.dword_5d4594_1316476)
		source := internCStr(lightningDrawSource)
		dx := num * nox_common_randomIntMinMax_415FF0(-jitter, jitter, source, 193) / den
		dy := num * nox_common_randomIntMinMax_415FF0(-jitter, jitter, source, 196) / den
		mid := lightningPackPoint(C.int2{
			field_0: C.int(dx) + ((from.field_0 + to.field_0) >> 1),
			field_4: C.int(dy) + ((from.field_4 + to.field_4) >> 1),
		})
		nox_xxx_drawLightningStep_4BB070(start, mid)
		nox_xxx_drawLightningStep_4BB070(mid, end)
	}
	legacyWords.dword_5d4594_1316492--
	return C.int(legacyWords.dword_5d4594_1316492)
}

func lightningConfigureDistance(dx, dy int32) {
	var distance int32
	if squared := dx*dx + dy*dy; squared > 0 {
		distance = int32(math.Sqrt(float64(squared)))
	}
	steps := legacyWords.nox_xxx_lightningSteps_587000_178216
	far := memmap.Uint32(0x587000, 178204)
	if distance >= 512 {
		legacyWords.dword_5d4594_1316476 = C.uint32_t(far)
		legacyWords.dword_5d4594_1316448 = steps
		return
	}
	near := memmap.Uint32(0x587000, 178208)
	legacyWords.dword_5d4594_1316476 = C.uint32_t(near + uint32(distance)*(far-near)/512)
	switch {
	case distance < 64:
		steps = C.uint32_t(max(1, int32(steps)-3))
	case distance < 128:
		steps = C.uint32_t(max(1, int32(steps)-2))
	case distance < 256:
		steps = C.uint32_t(max(1, int32(steps)-1))
	}
	legacyWords.dword_5d4594_1316448 = steps
}

//export nox_xxx_lightningProc2_4BAE60
func nox_xxx_lightningProc2_4BAE60(from, to *C.int2, mode C.int, bounds *C.short, _ C.int, sideLayers, core C.int) C.int {
	dx, dy := lightningSpan(*from, *to)
	lightningConfigureDistance(dx, dy)
	*memmap.PtrUint32(0x5D4594, 1316532) = uint32(mode)
	if mode == 1 || mode == 3 {
		limit := unsafe.Slice(bounds, 3)
		*memmap.PtrUint16(0x5D4594, 1316432) = uint16(limit[0])
		*memmap.PtrUint16(0x5D4594, 1316434) = uint16(limit[2])
		var axis int
		if dx <= dy {
			axis = 2 + bool2int(from.field_4 >= to.field_4)
		} else {
			axis = bool2int(from.field_0 >= to.field_0)
		}
		*memmap.PtrUint32(0x5D4594, 1316500) = uint32(axis)
	}

	start, end := lightningPackPoint(*from), lightningPackPoint(*to)
	if sideLayers != 0 {
		*memmap.PtrUint32(0x5D4594, 1316508) = 0
		for _, color := range [...]C.uint32_t{legacyWords.dword_5d4594_1316456, legacyWords.dword_5d4594_1316452} {
			legacyWords.dword_5d4594_1316492 = 1
			legacyWords.dword_5d4594_1316472 = color
			nox_xxx_drawLightningStep_4BB070(start, end)
		}
	}
	if core != 0 {
		legacyWords.dword_5d4594_1316492 = 1
		legacyWords.dword_5d4594_1316472 = legacyWords.dword_5d4594_1316436
		*memmap.PtrUint32(0x5D4594, 1316440) = uint32(legacyWords.dword_5d4594_1316484)
		*memmap.PtrUint32(0x5D4594, 1316508) = 1
		return nox_xxx_drawLightningStep_4BB070(start, end)
	}
	return core
}

type lightningKind uint8

const (
	lightningNormal lightningKind = iota
	lightningEnergy
	lightningGreen
)

func drawLightningRay(view *C.int, dr *nox_drawable, kind lightningKind) C.int {
	ray := drawableRayPayload(asDrawable(dr))
	if kind == lightningGreen && ray[0] == 0 {
		if ticks := binary.LittleEndian.Uint32(ray[1:5]); ticks != 0 {
			ticks--
			binary.LittleEndian.PutUint32(ray[1:5], ticks)
			if ticks == 0 {
				nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
				return 0
			}
		}
	}
	from, to, ok := drawableRayEndpoints(ray)
	if !ok {
		return 1
	}
	vp := (*noxrender.Viewport)(unsafe.Pointer(view))
	p1, p2 := drawableRayScreen(vp, from), drawableRayScreen(vp, to)
	var particleOffset uintptr
	width, sideLayers := byte(1), C.int(1)
	switch kind {
	case lightningNormal:
		legacyWords.dword_5d4594_1316452 = C.uint32_t(memmap.Uint32(0x5D4594, 1316428))
		legacyWords.dword_5d4594_1316436 = C.uint32_t(memmap.Uint32(0x5D4594, 1316464))
		legacyWords.dword_5d4594_1316456 = C.uint32_t(memmap.Uint32(0x5D4594, 1316424))
		legacyWords.dword_5d4594_1316484 = C.uint32_t(memmap.Uint32(0x5D4594, 1316488))
		particleOffset = 1316520
	case lightningEnergy:
		width, sideLayers = byte(2*(int(int8(ray[1]))+127)), 0
		legacyWords.dword_5d4594_1316436 = C.uint32_t(memmap.Uint32(0x5D4594, 1316496))
		legacyWords.dword_5d4594_1316484 = C.uint32_t(memmap.Uint32(0x5D4594, 1316468))
		particleOffset = 1316524
	case lightningGreen:
		legacyWords.dword_5d4594_1316452 = C.uint32_t(memmap.Uint32(0x5D4594, 1316444))
		legacyWords.dword_5d4594_1316436 = C.uint32_t(memmap.Uint32(0x5D4594, 1316504))
		legacyWords.dword_5d4594_1316456 = C.uint32_t(memmap.Uint32(0x5D4594, 1316460))
		legacyWords.dword_5d4594_1316484 = C.uint32_t(memmap.Uint32(0x5D4594, 1316480))
		particleOffset = 1316528
	}
	*memmap.PtrUint8(0x5D4594, 1316420) = width
	nox_xxx_lightningProc2_4BAE60(&p1, &p2, 2, nil, sideLayers, sideLayers, 1)
	if nox_xxx_checkGameFlagPause_413A50() == 0 {
		nox_xxx_makeLightningParticles_4999D0(C.int(memmap.Int32(0x5D4594, particleOffset)), &from, &to)
	}
	return 1
}

//export nox_thing_lightning_draw
func nox_thing_lightning_draw(view *C.int, dr *nox_drawable) C.int {
	return drawLightningRay(view, dr, lightningNormal)
}

//export nox_thing_chain_lightning_bolt_draw
func nox_thing_chain_lightning_bolt_draw(view *C.int, dr *nox_drawable) C.int {
	return drawLightningRay(view, dr, lightningNormal)
}

//export nox_thing_energy_bolt_draw
func nox_thing_energy_bolt_draw(view *C.int, dr *nox_drawable) C.int {
	return drawLightningRay(view, dr, lightningEnergy)
}

//export nox_thing_green_bolt_draw
func nox_thing_green_bolt_draw(view *C.int, dr *nox_drawable) C.int {
	return drawLightningRay(view, dr, lightningGreen)
}

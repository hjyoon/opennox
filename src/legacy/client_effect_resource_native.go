package legacy

/*
#include "GAME2_3.h"

*/
import "C"

import (
	"encoding/binary"
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var reflectiveShieldTypes = [...]string{
	"ReflectiveShieldNW", "ReflectiveShieldN", "ReflectiveShieldNE",
	"ReflectiveShieldW", "", "ReflectiveShieldE",
	"ReflectiveShieldSW", "ReflectiveShieldS", "ReflectiveShieldSE",
}

func reflectiveShieldSlot(direction int) **nox_drawable {
	return (**nox_drawable)(unsafe.Pointer(memmap.PtrPtr(0x5D4594, 1217468+4*uintptr(direction))))
}

//export nox_xxx_loadReflSheild_499360
func nox_xxx_loadReflSheild_499360() C.int {
	for direction, name := range reflectiveShieldTypes {
		var shield *nox_drawable
		if name != "" {
			shield = nox_new_drawable_for_thing(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(name)))
			if shield == nil {
				return 0
			}
			shield.flags30 |= 0x1000000
		}
		*reflectiveShieldSlot(direction) = shield
	}
	*g2ServerWord(1217504) = 0
	return 1
}

//export sub_499450
func sub_499450() C.int {
	var result C.int
	for direction := range reflectiveShieldTypes {
		slot := reflectiveShieldSlot(direction)
		if shield := *slot; shield != nil {
			result = C.int(nox_xxx_spriteDelete_45A4B0(shield))
		}
		*slot = nil
	}
	*g2ServerWord(1217504) = 0
	return result
}

//export nox_xxx_drawShield_499810
func nox_xxx_drawShield_499810(view *nox_draw_viewport_t, drawable *nox_drawable) C.int {
	d := asDrawable(drawable)
	direction := int(d.AnimDir)
	if direction >= len(reflectiveShieldTypes) {
		return 0
	}
	shield := asDrawable(*reflectiveShieldSlot(direction))
	if shield != nil {
		offsets := memmap.PtrT[[9][2]int32](0x587000, 161776)
		at := offsets[direction]
		shield.SetPos(d.Pos().Add(image.Pt(int(at[0]), int(d.ZVal)+int(at[1]))))
		shield.CallDraw(asViewport(view))
	}
	return 0
}

//export nox_xxx_bookRewardCli_499CF0
func nox_xxx_bookRewardCli_499CF0(kind C.uintptr_t, id, extra C.int) {
	last := g2ServerWord(1217504)
	if noxflags.HasGame(noxflags.GameModeCoop) && uint32(nox_xxx_bookGet_430B40_get_mouse_prev_seq())-*last < 2 {
		return
	}
	*last = uint32(nox_xxx_bookGet_430B40_get_mouse_prev_seq())
	color := C.int(2)
	switch kind {
	case 2:
		color = 0
	case 3:
		color = 3
	}
	at := C.int2{field_0: 5, field_4: clientDisplayStorage.height / 3}
	nox_xxx_bookSetForward_45D200(kind, id, &at)
	x, y := at.field_0, at.field_4
	nox_xxx_draw_499E70(color, x, y, 271, 166, 1, 1)
	nox_xxx_draw_499E70(color, x, y, 135, 166, 2, 1)
	nox_xxx_draw_499E70(color, x, y+166, 135, 166, 2, 1)
	nox_xxx_draw_499E70(color, x+271, y, 271, 166, 1, 2)
	nox_xxx_draw_499E70(color, x+135, y, 135, 166, 2, 2)
	nox_xxx_draw_499E70(color, x+135, y+166, 135, 166, 2, 2)
	if kind != 4 && extra == 1 {
		nox_xxx_bookFillAll_45D570(C.int(kind), id)
	}
}

var bubbleParticleTypes = [...]struct {
	name   string
	center [3]byte
}{
	{"RedBubbleParticle", [3]byte{255, 128, 128}},
	{"WhiteBubbleParticle", [3]byte{255, 255, 255}},
	{"LightBlueBubbleParticle", [3]byte{200, 200, 255}},
	{"OrangeBubbleParticle", [3]byte{255, 100, 50}},
	{"GreenBubbleParticle", [3]byte{64, 255, 64}},
	{"VioletBubbleParticle", [3]byte{255, 100, 255}},
	{"LightVioletBubbleParticle", [3]byte{255, 200, 255}},
	{"YellowBubbleParticle", [3]byte{255, 255, 200}},
}

//export sub_499F60
func sub_499F60(typeID, x, y C.int, z C.short, radius, interval, mode, fade, extra C.char, lifetime C.int) {
	types := memmap.PtrT[[len(bubbleParticleTypes)]uint32](0x5D4594, 1217512)
	if types[0] == 0 {
		for i, entry := range bubbleParticleTypes {
			types[i] = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(entry.name)))
		}
	}
	dr := nox_xxx_spriteLoadAdd_45A360_drawable(int32(typeID), int32(x), int32(y))
	if dr == nil {
		return
	}
	d := asDrawable(dr)
	d.ZVal = uint16(z)
	light := d.LightColor
	base := noxcolor.RGB5551Color(byte(light.R), byte(light.G), byte(light.B))
	center := bubbleParticleTypes[2].center
	for i, id := range types {
		if i != 2 && id == uint32(typeID) {
			center = bubbleParticleTypes[i].center
			break
		}
	}
	payload := drawableEffectBytes(d)
	binary.LittleEndian.PutUint32(payload[:4], base.Color32())
	binary.LittleEndian.PutUint32(payload[4:8], noxcolor.RGB5551Color(center[0], center[1], center[2]).Color32())
	payload[8], payload[9] = byte(radius), 1
	payload[10], payload[11] = byte(interval), byte(interval)
	payload[12], payload[13], payload[14] = byte(fade), byte(extra), byte(mode)
	nox_xxx_spriteToSightDestroyList_49BAB0_drawable(dr)
	nox_xxx_spriteTransparentDecay_49B950(dr, int32(lifetime))
	nox_xxx_sprite_45A110_drawable(dr)
}

//export nox_xxx_fxDrawTurnUndead_499880
func nox_xxx_fxDrawTurnUndead_499880(point *C.short) *C.uint32_t {
	coords := (*[2]C.short)(unsafe.Pointer(point))
	pos := image.Pt(int(coords[0]), int(coords[1]))
	typeID := g2ServerWord(1217508)
	if *typeID == 0 {
		*typeID = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("UndeadKiller")))
	}
	var last *nox_drawable
	for direction := 0; direction < 256; direction += 6 {
		last = nox_xxx_spriteLoadAdd_45A360_drawable(int32(*typeID), int32(pos.X), int32(pos.Y))
		if last != nil {
			cosine, sine := server.SinCosDir(byte(direction))
			asDrawable(last).SetTurnUndeadEffect(byte(direction), pos, cosine, sine, gameFrame(), Get_nox_xxx_sprite_4CA540())
			nox_xxx_spriteToList_49BC80_drawable(last)
			nox_xxx_spriteToSightDestroyList_49BAB0_drawable(last)
		}
	}
	return (*C.uint32_t)(unsafe.Pointer(last))
}

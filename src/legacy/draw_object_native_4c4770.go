package legacy

/*
#include "GAME1_1.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__video__draw_common.h"
*/
import "C"

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var drawObjectPrevious struct {
	address uintptr
	z       int32
}

func drawObjectWallVisible4C4770(vp *noxrender.Viewport, dr *client.Drawable, worldDX int32) bool {
	if memmap.Uint32(0x587000, 80808) == 0 || dr.Flags()&0x40000001 != 0 || dr.Class()&0x80 == 0 {
		return true
	}
	from := C.int2{field_0: C.int(int32(vp.Screen.Min.X) + worldDX),
		field_4: C.int(int32(vp.Screen.Min.Y) + int32(dr.Pos().Y) - int32(vp.World.Min.Y))}
	direction := dr.Field_74_4
	offset := 196184 + 8*uintptr(direction)
	to := C.int2{
		field_0: from.field_0 + C.int(memmap.Int32(0x587000, offset)),
		field_4: from.field_4 + C.int(memmap.Int32(0x587000, offset+4)),
	}
	x, y := int32(from.field_0+to.field_0)>>1, int32(from.field_4+to.field_4)>>1
	if direction != 0 && direction < 24 && (direction < 8 || direction > 16) {
		if sub_4C5630(x-5, x-5, y) != 0 {
			from.field_0 -= 2
			to.field_0 -= 2
		} else if sub_4C5630(x+5, x+5, y) != 0 {
			from.field_0 += 2
			to.field_0 += 2
		}
	} else if sub_4C5630(x, x, y-5) != 0 {
		from.field_4 -= 2
		to.field_4 -= 2
	} else if sub_4C5630(x, x, y+5) != 0 {
		from.field_4 += 2
		to.field_4 += 2
	}
	var clipped int32
	cutoff := int32(clientDisplayStorage.width)
	return sub_4C42A0(&from, &to, &clipped, &cutoff) != 0
}

func drawObjectLight4C4770(dr *client.Drawable) *[9]byte {
	var light unsafe.Pointer
	if (dr.Class()&0x80000 != 0 && dr.Flags()&0x1000000 != 0) || dr.Flags()&0x40000000 != 0 {
		light = memmap.PtrOff(0x587000, 185472)
	} else {
		raw := dr.Pos()
		pos := C.nox_point{x: C.int(raw.X), y: C.int(raw.Y)}
		light = unsafe.Pointer(C.sub_469920(&pos))
	}
	return (*[9]byte)(light)
}

func drawObjectSetAlpha(alpha byte) {
	nox_client_drawEnableAlpha_434560(1)
	nox_client_drawSetAlpha_434580(C.uchar(alpha))
}

func drawObjectSetTint(color C.int) {
	nox_xxx_draw_434600(1)
	C.nox_draw_setColorMultAndIntensity_433E40(color)
}

func drawObjectFade4C4770(vp *noxrender.Viewport, dr *client.Drawable) {
	ghost := dr.Field27() == uint32(legacyWords.dword_5d4594_1321520)
	if dr.Field_120 != 0 {
		alpha := 1 - float64(gameFrame()-dr.Field_85)/float64(int32(gameFPS()))
		if alpha < 0 {
			alpha = 0.001
		}
		fade := byte(alpha * 255)
		if ghost {
			fade = min(fade, byte(sub_4C4EC0((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))))
		} else if dr.Flags()&0x4000000 != 0 {
			fade >>= 1
		}
		drawObjectSetAlpha(fade)
	} else if ghost {
		drawObjectSetAlpha(byte(sub_4C4EC0((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))))
	} else if dr.Flags()&0x4000000 != 0 {
		drawObjectSetAlpha(0x80)
	}
}

func drawObjectInvisibility4C4770(dr, local *client.Drawable, player *server.Player, sameTeam, observer bool) bool {
	if !dr.HasEnchant(server.ENCHANT_INVISIBLE) && !observer &&
		!(dr.Field27() == uint32(legacyWords.dword_5d4594_1321520) && local.HasEnchant(server.ENCHANT_INFRAVISION)) {
		return true
	}
	if player != nil && player.Field3680&1 != 0 {
		drawObjectSetAlpha(0x80)
		return true
	}
	pos := dr.Pos()
	dx, dy := int32(pos.X)-int32(dr.Field_8), int32(pos.Y)-int32(dr.Field_9)
	if dx > -4 && dx < 4 && dy > -4 && dy < 4 && local != nil && !local.HasEnchant(server.ENCHANT_INFRAVISION) && !sameTeam {
		return false
	}
	var distance int32
	if dx != 0 || dy != 0 {
		scale := int32(dr.Field_5) - int32(dr.Field_10)
		if scale == 0 {
			scale = 1
		}
		distance = int32(nox_double2int(C.double(math.Sqrt(float64(dx*dx+dy*dy))))) / scale
	}
	if observer {
		drawObjectSetAlpha(0x80)
		return true
	}
	if !sameTeam && local.HasEnchant(server.ENCHANT_INFRAVISION) {
		drawObjectSetTint(C.int(worldWordStorage.local_player_color))
		drawObjectSetAlpha(0xff)
		return true
	}
	alpha := byte(0x80)
	if distance < 8 {
		alpha = byte((distance << 7) / 8)
	}
	if alpha == 0 {
		alpha = 1
	}
	if sameTeam && alpha <= 1 &&
		((dr.Class()&0x2 != 0 && dr.AnimInd == 8) || (dr.Class().Has(object.ClassPlayer) && dr.AnimInd == 0)) {
		drawObjectSetTint(C.int(worldWordStorage.local_player_color))
		alpha = 0x80
	}
	drawObjectSetAlpha(alpha)
	return true
}

//export nox_xxx_drawObject_4C4770_draw
func nox_xxx_drawObject_4C4770_draw(vp *nox_draw_viewport_t, dr *nox_drawable, img unsafe.Pointer) {
	drawObject4C4770((*noxrender.Viewport)(unsafe.Pointer(vp)), asDrawable(dr), noxrender.ImageHandle(img))
}

func drawObject4C4770(vp *noxrender.Viewport, dr *client.Drawable, img noxrender.ImageHandle) {
	if vp == nil || dr == nil || img == nil {
		return
	}
	if legacyWords.dword_5d4594_1321520 == 0 {
		legacyWords.dword_5d4594_1321520 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("Ghost")))
	}
	var sameTeam, observer bool
	if dr.Class().Has(object.ClassPlayer) {
		localCode := uint32(ClientPlayerNetCode())
		info := asPlayerS(nox_common_playerInfoGetByID_417040(int32(dr.Field32())))
		if info != nil && info.Field3680&1 != 0 {
			if localCode != dr.Field32() {
				return
			}
			observer = true
		}
		local := nox_xxx_objGetTeamByNetCode_418C80(C.int(localCode))
		team := nox_xxx_objGetTeamByNetCode_418C80(C.int(dr.Field32()))
		sameTeam = local != nil && team != nil &&
			(localCode == dr.Field32() || nox_xxx_servCompareTeams_419150(local, team) != 0)
	}

	pos := dr.Pos()
	worldDX := int32(pos.X) - int32(vp.World.Min.X)
	screenX := worldDX + int32(vp.Screen.Min.X) - int32(byte(dr.Field_0))
	screenY := int32(vp.Screen.Min.Y) + int32(pos.Y) - int32(int16(dr.ZVal)) -
		int32(byte(dr.Field_0>>8)) - int32(int16(dr.ZVal2)) - int32(vp.World.Min.Y)
	if !drawObjectWallVisible4C4770(vp, dr, worldDX) {
		return
	}

	light := drawObjectLight4C4770(dr)
	z := int32(int16(dr.ZVal))
	address := uintptr(unsafe.Pointer(dr))
	restoreClip := z < 0 && drawObjectPrevious.z < 0 && drawObjectPrevious.address == address
	drawObjectPrevious.address, drawObjectPrevious.z = address, z
	if dr.HasEnchant(server.ENCHANT_FREEZE) ||
		(dr.Class()&0x2 != 0 && dr.Flags()&0x40000000 != 0 && dr.Flags()&0x8020 == 0) {
		drawObjectSetTint(C.int(renderPalette.blue))
	} else {
		sub_4345F0(1)
		C.nox_draw_setColorMultAndIntensityRGB_433CD0(C.uint8_t(light[0]), C.uint8_t(light[4]), C.uint8_t(light[8]))
	}

	drawObjectFade4C4770(vp, dr)
	if !drawObjectInvisibility4C4770(dr, AsDrawableP(*memmap.PtrPtr(0x852978, 8)), Get_dword_8531A0_2576(), sameTeam, observer) {
		return
	}
	if !dr.Class().Has(object.ClassPlayer) && dr.HasEnchant(server.ENCHANT_INVULNERABLE) && !noxflags.HasGame(noxflags.GameModeCoop) {
		tint := renderPalette.blue
		if gameFrame()&1 != 0 {
			tint = renderPalette.white
		}
		drawObjectSetTint(C.int(tint))
	}
	if !dr.HasEnchant(server.ENCHANT_INVULNERABLE) && dr.HasEnchant(server.ENCHANT_AFRAID) {
		drawObjectSetTint(C.int(memmap.Int32(0x85B3FC, 956)))
	}
	if restoreClip {
		nox_xxx_wndDraw_49F7F0()
		C.sub_49F7C0_def()
	} else {
		C.sub_47D370(C.int(max(0, -z)))
	}
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(img), screenX, screenY)
	sub_4345F0(0)
	nox_client_drawEnableAlpha_434560(0)
	nox_xxx_draw_434600(0)
	if restoreClip {
		sub_49F860()
	}
	dr.Field_0 = dr.Field_0&0xffff | 1<<16
	dr.Field_1 = uint32(memmap.Uint16(0x973F18, 88)) | uint32(memmap.Uint16(0x973F18, 76))<<16
	dr.Field_2 = unsafe.Pointer(img)
}

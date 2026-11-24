package legacy

/*
#include "GAME2_2.h"

extern uint32_t dword_5d4594_1096636;
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_packetGetMarshall_476F40
func nox_xxx_packetGetMarshall_476F40() C.uint {
	return C.uint(objectFieldDrawableCode578B00(Get_dword_5d4594_1096640()))
}

//export nox_xxx_clientEnumHover_476FA0
func nox_xxx_clientEnumHover_476FA0() {
	if glyph := g2ServerWord(1096632); *glyph == 0 {
		*glyph = uint32(GetServer().S().Types.IndByID("Glyph"))
	}
	cli := GetClient()
	pos := Sub_473970(cli.GetMousePos())
	Set_dword_5d4594_1096640(nil)
	Set_nox_client_spriteUnderCursorXxx_1096644(nil)
	*g2ServerWord(1096628) = 0
	C.dword_5d4594_1096636 = 0
	bounds := image.Rect(pos.X-96, pos.Y-96, pos.X+96, pos.Y+96)
	cli.Cli().Objs.EachInRect(bounds, func(dr *client.Drawable) {
		clientOnCursorHover(dr, &pos)
	})
}

//export nox_xxx_clientOnCursorHover_477050
func nox_xxx_clientOnCursorHover_477050(dr *nox_drawable, data unsafe.Pointer) {
	if data == nil {
		clientOnCursorHover(asDrawable(dr), nil)
		return
	}
	pt := (*C.nox_point)(data)
	pos := image.Pt(int(pt.x), int(pt.y))
	clientOnCursorHover(asDrawable(dr), &pos)
}

func clientOnCursorHover(dr *client.Drawable, mouse *image.Point) {
	polyp := g2ServerWord(1096648)
	if *polyp == 0 {
		*polyp = uint32(GetClient().Cli().Things.IndByID("Polyp"))
	}
	local := AsDrawableP(*memmap.PtrPtr(0x852978, 8))
	if dr == nil || mouse == nil || dr == local {
		return
	}
	if dr.ObjFlags&0x8000 != 0 || (dr.HasEnchant(0) && !local.HasEnchant(21)) {
		return
	}
	class, subclass := dr.ObjClass, dr.ObjSubClass
	if class&2 != 0 && subclass&0x4000 != 0 {
		return
	}
	if class&0x80400206 == 0 && dr.TypeIDVal != *polyp {
		return
	}
	if !GetClient().Nox_xxx_client_4984B0_drawable(dr) {
		return
	}
	if class&4 != 0 {
		info := nox_common_playerInfoGetByID_417040(int32(dr.NetCode32))
		if info == nil || info.field_3680&1 != 0 {
			return
		}
	}
	if (class&0x400000 != 0 && subclass&0x80 == 0) || (class&2 != 0 && dr.AnimInd == 10) {
		return
	}
	if !cursorHoverContains(dr, *mouse) {
		return
	}

	depth := server.X87TruncateSignedDWORD419A70(float32(int16(dr.ZVal)) + float32(dr.PosVec.Y) + dr.ZSizeMin)
	if top := g2ServerWord(1096628); depth > int32(*top) {
		*top = uint32(depth)
		Set_dword_5d4594_1096640(dr)
	}
	if depth <= int32(C.dword_5d4594_1096636) ||
		nox_xxx_client_57B400((*nox_drawable)(dr.C())) == 0 {
		return
	}
	info := Get_dword_8531A0_2576()
	if info != nil && info.PlayerClass() == player.Wizard && dr.TypeIDVal == *g2ServerWord(1096632) {
		if Get_nox_client_spriteUnderCursorXxx_1096644() == nil {
			Set_nox_client_spriteUnderCursorXxx_1096644(dr)
			C.dword_5d4594_1096636 = 0
		}
	} else {
		C.dword_5d4594_1096636 = C.uint32_t(depth)
		Set_nox_client_spriteUnderCursorXxx_1096644(dr)
	}
}

func cursorHoverContains(dr *client.Drawable, mouse image.Point) bool {
	toInt := server.X87TruncateSignedDWORD419A70
	z := float32(int16(dr.ZVal))
	cx, cy := int32(dr.PosVec.X), float32(dr.PosVec.Y)
	x, y := int32(mouse.X), int32(mouse.Y)
	lower := toInt(cy - dr.ZSizeMax - z)
	upper := toInt(cy - dr.ZSizeMin - z)
	switch dr.Shape.Kind {
	case server.ShapeKindCircle:
		r := dr.Shape.Circle.R
		radius := toInt(r)
		if y >= lower && y <= upper {
			return x > cx-radius && x < cx+radius
		}
		dx, dy := x-cx, y-upper
		return dx*dx+dy*dy < toInt(r*r)
	case server.ShapeKindBox:
		cursor := types.Pointf{X: float32(x), Y: float32(y)}
		origin := types.Pointf{X: float32(dr.PosVec.X), Y: float32(upper)}
		if server.MapPointInBox57B850(&origin, &dr.Shape, &cursor) {
			return true
		}
		origin.Y = float32(lower)
		if server.MapPointInBox57B850(&origin, &dr.Shape, &cursor) {
			return true
		}
		box := &dr.Shape.Box
		left, leftY := toInt(box.LeftBottom2), toInt(box.LeftTop2)
		if x > cx+left && x < cx && y > lower+leftY && y < upper+leftY {
			return true
		}
		right, rightY := toInt(box.RightTop), toInt(box.RightBottom)
		return x >= cx && x < cx+right && y > lower+rightY && y < upper+rightY
	default:
		return false
	}
}

//export nox_xxx_cursorSetTooltip_4776B0
func nox_xxx_cursorSetTooltip_4776B0(text *wchar2_t) {
	dst := memmap.PtrT[[256]uint16](0x5D4594, 1096676)
	if text == nil {
		dst[0] = 0
		return
	}
	for i := range len(dst) - 1 {
		word := *(*uint16)(unsafe.Add(unsafe.Pointer(text), i*2))
		dst[i] = word
		if word == 0 {
			return
		}
	}
	dst[len(dst)-1] = 0
}

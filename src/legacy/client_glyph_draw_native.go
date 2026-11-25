package legacy

/*
#include "client__draw__glyphdraw.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type GlyphDrawState struct {
	Visible  bool
	Alpha    byte
	Colorize bool
}

func GlyphDrawStateFor(dr, local *client.Drawable, gameClient bool, backbufferDepth uint32) GlyphDrawState {
	if dr == nil {
		return GlyphDrawState{}
	}
	// The original x86 routine reused the low byte of the drawable pointer outside client mode.
	state := GlyphDrawState{Visible: true, Alpha: byte(uintptr(unsafe.Pointer(dr)))}
	if !gameClient || local == nil {
		return state
	}
	if dr.ObjFlags&0x40000000 != 0 {
		state.Alpha = 255
		return state
	}
	if local.HasEnchant(server.ENCHANT_INFRAVISION) {
		state.Colorize = true
		if backbufferDepth >= 16 {
			state.Alpha = 255
		} else {
			state.Alpha = 128
		}
		return state
	}
	dx := int64(dr.PosVec.X) - int64(local.PosVec.X)
	dy := int64(dr.PosVec.Y) - int64(local.PosVec.Y)
	dist2 := dx*dx + dy*dy
	if dist2 >= 22500 {
		state.Visible = false
		return state
	}
	state.Alpha = byte(-56 - 200*dist2/22500)
	return state
}

//export nox_thing_glyph_draw
func nox_thing_glyph_draw(view *C.int, drawable *C.nox_drawable) C.int {
	local := AsDrawableP(*memmap.PtrPtr(0x852978, 8))
	state := GlyphDrawStateFor(asDrawable(drawable), local, noxflags.HasGame(noxflags.GameClient), 32)
	if !state.Visible {
		return 1
	}
	if state.Colorize {
		nox_xxx_draw_434600(1)
		nox_draw_setColorMultAndIntensity_433E40(int32(Get_dword_8531A0_2572()))
	}
	nox_client_drawEnableAlpha_434560(1)
	nox_client_drawSetAlpha_434580(C.uchar(state.Alpha))
	result := nox_thing_animate_draw((*C.uint32_t)(unsafe.Pointer(view)), drawable)
	nox_client_drawEnableAlpha_434560(0)
	nox_xxx_draw_434600(0)
	return result
}

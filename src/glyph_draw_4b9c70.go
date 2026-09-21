//go:build !server

package opennox

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type glyphDrawState4B9C70 struct {
	visible  bool
	alpha    byte
	colorize bool
}

func glyphDrawStateFor4B9C70(dr, local *client.Drawable, gameClient bool, backbufferDepth uint32) glyphDrawState4B9C70 {
	if dr == nil {
		return glyphDrawState4B9C70{}
	}
	// The original x86 routine reused the low byte of its drawable argument
	// when invoked outside client mode. That path is not used for an active
	// client, but retaining it keeps the translated routine deterministic with
	// GAME.EXE rather than inventing a new fallback alpha.
	state := glyphDrawState4B9C70{visible: true, alpha: byte(uintptr(unsafe.Pointer(dr)))}
	if !gameClient || local == nil {
		return state
	}
	if dr.ObjFlags&0x40000000 != 0 {
		state.alpha = 255
		return state
	}
	if local.HasEnchant(server.ENCHANT_INFRAVISION) {
		state.colorize = true
		if backbufferDepth >= 16 {
			state.alpha = 255
		} else {
			state.alpha = 128
		}
		return state
	}
	dx := int64(dr.PosVec.X - local.PosVec.X)
	dy := int64(dr.PosVec.Y - local.PosVec.Y)
	dist2 := dx*dx + dy*dy
	if dist2 >= 22500 {
		state.visible = false
		return state
	}
	state.alpha = byte(-56 - 200*dist2/22500)
	return state
}

func (c *Client) drawGlyph4B9C70(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := glyphDrawStateFor4B9C70(
		dr,
		c.ClientPlayerUnit(),
		noxflags.HasGame(noxflags.GameClient),
		// The native renderer is RGBA32. The original executable also had an
		// 8-bit fallback here, but OpenNox has no indexed-color backbuffer.
		32,
	)
	if !state.visible {
		return 1
	}
	draw := c.r.Data()
	if state.colorize {
		draw.SetColorize17(1)
		c.r.SetColorMultAndIntensity(nox_color_green)
	}
	draw.SetAlphaEnabled(true)
	draw.SetAlpha(state.alpha)
	res := legacy.Nox_thing_animate_draw(vp, dr)
	draw.SetAlphaEnabled(false)
	draw.SetColorize17(0)
	return res
}

func (c *Client) callGlyphDraw4B9C70(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_glyph_draw() {
		return 0, false
	}
	return c.drawGlyph4B9C70(dr, vp), true
}

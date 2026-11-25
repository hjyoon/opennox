//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

func (c *Client) drawGlyph4B9C70(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := legacy.GlyphDrawStateFor(
		dr,
		c.ClientPlayerUnit(),
		noxflags.HasGame(noxflags.GameClient),
		// The native renderer is RGBA32. The original executable also had an
		// 8-bit fallback here, but OpenNox has no indexed-color backbuffer.
		32,
	)
	if !state.Visible {
		return 1
	}
	draw := c.r.Data()
	if state.Colorize {
		draw.SetColorize17(1)
		c.r.SetColorMultAndIntensity(nox_color_green)
	}
	draw.SetAlphaEnabled(true)
	draw.SetAlpha(state.Alpha)
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

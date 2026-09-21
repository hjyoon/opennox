//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

type sphericalShieldDrawHooks4B8020 struct {
	byCode  func(uint16) *client.Drawable
	move    func(*client.Drawable, int, int)
	animate func(*client.Drawable) int
	delete  func(*client.Drawable)
}

func drawSphericalShield4B8020(dr *client.Drawable, hooks sphericalShieldDrawHooks4B8020) int {
	if dr == nil {
		return 0
	}
	target := hooks.byCode(uint16(dr.UnionEffect().Field_108))
	if target == nil {
		hooks.delete(dr)
		return 0
	}
	hooks.move(dr, target.PosVec.X, target.PosVec.Y+3)
	return hooks.animate(dr)
}

func (c *Client) drawSphericalShield4B8020(dr *client.Drawable, vp *noxrender.Viewport) int {
	return drawSphericalShield4B8020(dr, sphericalShieldDrawHooks4B8020{
		byCode: c.Objs.ByNetCode,
		move:   c.Nox_xxx_updateSpritePosition_49AA90,
		animate: func(got *client.Drawable) int {
			return legacy.Nox_thing_animate_draw(vp, got)
		},
		delete: c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
	})
}

func (c *Client) callSphericalShieldDraw4B8020(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_spherical_shield_draw() {
		return 0, false
	}
	return c.drawSphericalShield4B8020(dr, vp), true
}

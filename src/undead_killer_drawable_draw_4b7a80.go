//go:build !server

package opennox

import (
	"image"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap/nox/blobdata"
	"github.com/opennox/opennox/v1/legacy"
)

var undeadKillerGlow4B7A80 = noxcolor.RGB5551Color(100, 100, 255)

type undeadKillerDrawHooks4B7A80 struct {
	frame    func() uint32
	random   func(minimum, maximum int) int
	spawn    func(typeID int, pos image.Point) *client.Drawable
	activate func(*client.Drawable)
	delete   func(*client.Drawable)
	drawGlow func(pos image.Point, color noxcolor.RGBA5551, inner, outer int)
}

func init() {
	legacy.Nox_thing_undead_killer_draw = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.drawUndeadKiller4B7A80(dr, vp)
	}
}

func undeadKillerDistance4B7A80(from image.Point, toX, toY uint32) uint32 {
	dx := int32(toX) - int32(from.X)
	dy := int32(toY) - int32(from.Y)
	// The original x86 routine performs these operations in 32 bits.
	squared := uint32(dx*dx + dy*dy)
	return blobdata.ApproxSqrt32(squared)
}

func undeadKillerAlive4B7A80(frame, started uint32) bool {
	return frame-started <= 70
}

func spawnUndeadKillerOrb4B7A80(source *client.Drawable, whiteOrbType int, hooks undeadKillerDrawHooks4B7A80) {
	x := uint16(int32(uint16(source.PosVec.X)) + int32(hooks.random(-5, 5)))
	y := uint16(int32(uint16(source.PosVec.Y)) + int32(hooks.random(-5, 5)))
	size := byte(hooks.random(6, 10))
	yOffset := hooks.random(-5, 5)

	orb := hooks.spawn(whiteOrbType, image.Pt(int(x), int(y)+yOffset))
	if orb == nil {
		return
	}
	effect := orb.UnionEffect()
	effect.Field_108 = uint32(uint16(source.Field_81)) | uint32(uint16(source.Field_82))<<16
	effect.Field_110 = effect.Field_110&0x00ffffff | uint32(size)<<24
	effect.Field_111 = effect.Field_111&0xff000000 | uint32(byte(hooks.random(3, 10)))
	hooks.activate(orb)
}

func drawUndeadKiller4B7A80(source *client.Drawable, vp *noxrender.Viewport, whiteOrbType int, hooks undeadKillerDrawHooks4B7A80) int {
	if !undeadKillerAlive4B7A80(hooks.frame(), source.AnimStart) {
		hooks.delete(source)
		return 0
	}
	if hooks.random(0, 100) > 85 {
		spawnUndeadKillerOrb4B7A80(source, whiteOrbType, hooks)
	}

	point := vp.ToScreenPos(source.PosVec).Add(image.Pt(0, -int(int16(source.ZVal2))-int(int16(source.ZVal))))
	radius := 8 - int(undeadKillerDistance4B7A80(source.PosVec, source.Field_81, source.Field_82))/40
	if radius < 0 {
		radius = 1
	}
	hooks.drawGlow(point, undeadKillerGlow4B7A80, radius, 12)
	return 1
}

func (c *Client) undeadKillerDrawHooks4B7A80() undeadKillerDrawHooks4B7A80 {
	return undeadKillerDrawHooks4B7A80{
		frame:    c.srv.Frame,
		random:   c.srv.Rand.Other.Int,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
		delete:   c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
		drawGlow: func(pos image.Point, color noxcolor.RGBA5551, inner, outer int) {
			c.r.DrawGlow(pos, color, inner, outer)
		},
	}
}

func (c *Client) drawUndeadKiller4B7A80(dr *client.Drawable, vp *noxrender.Viewport) int {
	if dr == nil || vp == nil {
		return 0
	}
	return drawUndeadKiller4B7A80(dr, vp, c.Things.IndByID("WhiteOrb"), c.undeadKillerDrawHooks4B7A80())
}

func (c *Client) callUndeadKillerDraw4B7A80(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_undead_killer_draw() {
		return 0, false
	}
	return c.drawUndeadKiller4B7A80(dr, vp), true
}

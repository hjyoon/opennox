//go:build !server

package opennox

import (
	"image"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

var undeadKillerGlow4B7A80 = noxcolor.RGB5551Color(100, 100, 255)

// undeadKillerSqrtTable4B7A80 is byte_587000[155956:156212]. Keeping this
// immutable copy next to the port makes the draw callback independent of the
// legacy data-blob initialization order.
var undeadKillerSqrtTable4B7A80 = [256]byte{
	0, 16, 22, 27, 32, 35, 39, 42, 45, 48, 50, 53, 55, 57, 59, 61,
	64, 65, 67, 69, 71, 73, 75, 76, 78, 80, 81, 83, 84, 86, 87, 89,
	90, 91, 93, 94, 96, 97, 98, 99, 101, 102, 103, 104, 106, 107, 108, 109,
	110, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126,
	128, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142,
	143, 144, 144, 145, 146, 147, 148, 149, 150, 150, 151, 152, 153, 154, 155, 155,
	156, 157, 158, 159, 160, 160, 161, 162, 163, 163, 164, 165, 166, 167, 167, 168,
	169, 170, 170, 171, 172, 173, 173, 174, 175, 176, 176, 177, 178, 178, 179, 180,
	181, 181, 182, 183, 183, 184, 185, 185, 186, 187, 187, 188, 189, 189, 190, 191,
	192, 192, 193, 193, 194, 195, 195, 196, 197, 197, 198, 199, 199, 200, 201, 201,
	202, 203, 203, 204, 204, 205, 206, 206, 207, 208, 208, 209, 209, 210, 211, 211,
	212, 212, 213, 214, 214, 215, 215, 216, 217, 217, 218, 218, 219, 219, 220, 221,
	221, 222, 222, 223, 224, 224, 225, 225, 226, 226, 227, 227, 228, 229, 229, 230,
	230, 231, 231, 232, 232, 233, 234, 234, 235, 235, 236, 236, 237, 237, 238, 238,
	239, 240, 240, 241, 241, 242, 242, 243, 243, 244, 244, 245, 245, 246, 246, 247,
	247, 248, 248, 249, 249, 250, 250, 251, 251, 252, 252, 253, 253, 254, 254, 255,
}

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

// undeadKillerSqrt4B7A80 is the integer square-root approximation used by
// GAME.EXE 0048C730. The lookup table is part of the original immutable data
// blob; mirroring the shifts preserves the exact glow-radius transitions.
func undeadKillerSqrt4B7A80(value uint32) uint32 {
	lookup := func(index uint32) uint32 {
		return uint32(undeadKillerSqrtTable4B7A80[index])
	}
	if value < 0x10000 {
		if value < 0x100 {
			if value < 0x10 {
				if value < 4 {
					return lookup(64*value) >> 7
				}
				return lookup(16*value) >> 6
			}
			if value < 0x40 {
				return lookup(4*value) >> 5
			}
			return lookup(value) >> 4
		}
		if value < 0x1000 {
			if value < 0x400 {
				return lookup(value>>2) >> 3
			}
			return lookup(value>>4) >> 2
		}
		if value < 0x4000 {
			return lookup(value>>6) >> 1
		}
		return lookup(value >> 8)
	}
	if value < 0x1000000 {
		if value < 0x100000 {
			if value < 0x40000 {
				return lookup(value>>10) << 1
			}
			return lookup(value>>12) << 2
		}
		if value < 0x400000 {
			return lookup(value>>14) << 3
		}
		return lookup(value>>16) << 4
	}
	if value < 0x10000000 {
		if value < 0x4000000 {
			return lookup(value>>18) << 5
		}
		return lookup(value>>20) << 6
	}
	if value < 0x40000000 {
		return lookup(value>>22) << 7
	}
	return lookup(value>>24) << 8
}

func undeadKillerDistance4B7A80(from image.Point, toX, toY uint32) uint32 {
	dx := int32(toX) - int32(from.X)
	dy := int32(toY) - int32(from.Y)
	// The original x86 routine performs these operations in 32 bits.
	squared := uint32(dx*dx + dy*dy)
	return undeadKillerSqrt4B7A80(squared)
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

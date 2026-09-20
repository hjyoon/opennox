package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client"
)

type vortexSourceDrawableHooks4CC950 struct {
	typeID       func(string) int
	random       func(min, max int) int
	spawn        func(typ int, pos image.Point) *client.Drawable
	dimColor     uint32
	brightColor  uint32
	activate     func(*client.Drawable)
	sightDestroy func(*client.Drawable)
}

// updateVortexSourceDrawable4CC950 replaces the PE32 callback, which passed
// the source drawable as an int and stored the spawned orb in another int.
func updateVortexSourceDrawable4CC950(source *client.Drawable, hooks vortexSourceDrawableHooks4CC950) int {
	orbType := hooks.typeID("WhiteVortexOrb")
	angle := byte(hooks.random(0, 255))
	direction := sincosTable16[angle]
	position := source.PosVec.Add(image.Pt(50*direction.X/16, 50*direction.Y/16))
	orb := hooks.spawn(orbType, position)
	if orb == nil {
		return 1
	}

	spin := byte(hooks.random(2, 3))
	if hooks.random(0, 100) > 50 {
		spin = -spin
	}
	effect := orb.UnionEffect()
	effect.Field_108 = hooks.dimColor
	effect.Field_109 = hooks.brightColor
	effect.Field_110 = uint32(source.PosVec.X)
	effect.Field_111 = uint32(source.PosVec.Y)
	effect.Field_112 = uint32(angle) | uint32(spin)<<8 | 50<<16 | 1<<24
	source.ZVal = 0
	hooks.activate(orb)
	hooks.sightDestroy(orb)
	return 1
}

func (c *Client) updateVortexSourceDrawable4CC950(source *client.Drawable) int {
	return updateVortexSourceDrawable4CC950(source, vortexSourceDrawableHooks4CC950{
		typeID:       c.Things.IndByID,
		random:       c.srv.Rand.Other.Int,
		spawn:        c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		dimColor:     nox_color_rgb_4344A0(200, 200, 200),
		brightColor:  nox_color_rgb_4344A0(255, 255, 255),
		activate:     c.Objs.List34Add,
		sightDestroy: c.Objs.List6Add,
	})
}

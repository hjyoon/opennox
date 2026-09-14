package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client"
)

type teleportWakeDrawableHooks4CD8D0 struct {
	typeID   func() int
	random   func(min, max int) int
	frame    func() uint32
	spawn    func(typ int, pos image.Point) *client.Drawable
	activate func(*client.Drawable)
}

// updateTeleportWakeDrawable4CD8D0 replaces the PE32 client update, whose
// int return from spriteLoadAdd truncates newly allocated drawable pointers.
func updateTeleportWakeDrawable4CD8D0(wake *client.Drawable, hooks teleportWakeDrawableHooks4CD8D0) int {
	pos := wake.PosVec.Add(image.Pt(hooks.random(-5, 5), hooks.random(-5, 5)))
	spark := hooks.spawn(hooks.typeID(), pos)
	if spark == nil {
		return 1
	}
	effect := spark.UnionEffect()
	effect.Field_108 = uint32(spark.PosVec.X) << 12
	effect.Field_109 = uint32(spark.PosVec.Y) << 12
	spark.Field_74_4 = byte(hooks.random(0, 255))
	effect.Field_110 = uint32(hooks.random(1, 100))
	effect.Field_112 = hooks.frame() + uint32(hooks.random(10, 32))
	effect.Field_111 = hooks.frame()
	spark.ZVal = 0
	spark.VelZ = int8(hooks.random(3, 8))
	hooks.activate(spark)
	return 1
}

func (c *Client) updateTeleportWakeDrawable4CD8D0(wake *client.Drawable) int {
	return updateTeleportWakeDrawable4CD8D0(wake, teleportWakeDrawableHooks4CD8D0{
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxPointSparkTypes[0], "BlueSpark", c.Things.IndByID,
			)
		},
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

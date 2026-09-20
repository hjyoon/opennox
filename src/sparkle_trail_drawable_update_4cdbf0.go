package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type sparkleTrailDrawableHooks4CDBF0 struct {
	typeID   func(string) int
	random   func(min, max int) int
	frame    func() uint32
	spawn    func(typ int, pos image.Point) *client.Drawable
	drawFunc unsafe.Pointer
	activate func(*client.Drawable)
}

// updateSparkleTrailDrawable4CDBF0 replaces the PE32 callback that stored the
// spriteLoadAdd result in an int before initializing each BlueSpark.
func updateSparkleTrailDrawable4CDBF0(trail *client.Drawable, hooks sparkleTrailDrawableHooks4CDBF0) int {
	prev := image.Pt(int(int32(trail.Field_8)), int(int32(trail.Field_9)))
	delta := trail.PosVec.Sub(prev)
	typ := hooks.typeID("BlueSpark")
	offset := image.Point{}
	for range 5 {
		pos := image.Pt(
			prev.X+offset.X/5+hooks.random(-3, 3),
			prev.Y+hooks.random(-3, 3)+offset.Y/5,
		)
		spark := hooks.spawn(typ, pos)
		if spark != nil {
			spark.DrawFuncPtr = hooks.drawFunc
			spark.SetLightColor(255, 200, 75)
			effect := spark.UnionEffect()
			effect.Field_108 = uint32(pos.X) << 12
			effect.Field_109 = uint32(pos.Y) << 12
			spark.Field_74_4 = 0
			effect.Field_110 = 0
			effect.Field_112 = hooks.frame() + uint32(hooks.random(2, 10))
			effect.Field_111 = hooks.frame()
			spark.ZVal = trail.ZVal
			spark.ZVal2 = trail.ZVal2
			spark.VelZ = 0
			hooks.activate(spark)
		}
		offset = offset.Add(delta)
	}
	return 1
}

func (c *Client) updateSparkleTrailDrawable4CDBF0(trail *client.Drawable) int {
	return updateSparkleTrailDrawable4CDBF0(trail, sparkleTrailDrawableHooks4CDBF0{
		typeID:   c.Things.IndByID,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		drawFunc: legacy.Get_nox_thing_pixie_dust_draw(),
		activate: c.Objs.List34Add,
	})
}

package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type magicDrawableHooks4CDD80 struct {
	typeID   func(string) int
	random   func(min, max int) int
	frame    func() uint32
	fps      func() uint32
	spawn    func(typ int, pos image.Point) *client.Drawable
	drawFunc unsafe.Pointer
	activate func(*client.Drawable)
	decay    func(*client.Drawable, int)
}

// updateMagicDrawable4CDD80 replaces the PE32 callback, which stored each
// BlueSpark returned by spriteLoadAdd in a 32-bit integer before using it.
func updateMagicDrawable4CDD80(magic *client.Drawable, hooks magicDrawableHooks4CDD80) int {
	effect := magic.UnionEffect()
	dx := uint32(magic.PosVec.X) - effect.Field_108
	dy := uint32(magic.PosVec.Y) - effect.Field_109
	if dx*dx+dy*dy > 200 {
		tailType := hooks.typeID("MagicTailLink")
		tail := hooks.spawn(tailType, image.Pt(
			int(int32(effect.Field_108)),
			int(int32(effect.Field_109)),
		))
		if tail != nil {
			tailEffect := tail.UnionEffect()
			tailEffect.Field_108 = uint32(magic.PosVec.X)
			tailEffect.Field_109 = uint32(magic.PosVec.Y)
			hooks.activate(tail)
			effect.Field_108 = uint32(magic.PosVec.X)
			effect.Field_109 = uint32(magic.PosVec.Y)
			hooks.decay(tail, int(hooks.fps()))
		}
	}

	prev := image.Pt(int(int32(magic.Field_8)), int(int32(magic.Field_9)))
	delta := magic.PosVec.Sub(prev)
	sparkType := hooks.typeID("BlueSpark")
	offset := image.Point{}
	for range 4 {
		pos := image.Pt(
			prev.X+offset.X/4+hooks.random(-8, 8),
			prev.Y+offset.Y/4+hooks.random(-8, 8),
		)
		spark := hooks.spawn(sparkType, pos)
		if spark != nil {
			spark.DrawFuncPtr = hooks.drawFunc
			spark.SetLightColor(128, 128, 255)
			sparkEffect := spark.UnionEffect()
			sparkEffect.Field_108 = uint32(magic.PosVec.X) << 12
			sparkEffect.Field_109 = uint32(magic.PosVec.Y) << 12
			spark.Field_74_4 = 0
			sparkEffect.Field_110 = 0
			sparkEffect.Field_112 = hooks.frame() + uint32(hooks.random(10, 20))
			sparkEffect.Field_111 = hooks.frame()
			spark.ZVal = magic.ZVal
			spark.ZVal2 = magic.ZVal2
			spark.VelZ = 0
			hooks.activate(spark)
		}
		offset = offset.Add(delta)
	}
	return 1
}

func (c *Client) updateMagicDrawable4CDD80(magic *client.Drawable) int {
	return updateMagicDrawable4CDD80(magic, magicDrawableHooks4CDD80{
		typeID:   c.Things.IndByID,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		fps:      c.srv.TickRate,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		drawFunc: legacy.Get_nox_thing_magic_sparkle_draw(),
		activate: c.Objs.List34Add,
		decay:    c.Objs.TransparentDecay,
	})
}

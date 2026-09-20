package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

type magicMissileDrawableHooks4CD9E0 struct {
	sparkCount func() int32
	typeID     func(string) int
	random     func(min, max int) int
	frame      func() uint32
	fps        func() uint32
	spawn      func(typ int, pos image.Point) *client.Drawable
	activate   func(*client.Drawable)
	decay      func(*client.Drawable, int)
}

// updateMagicMissileDrawable4CD9E0 replaces the PE32 callback, which stored
// both spriteLoadAdd results in 32-bit integers before dereferencing them.
func updateMagicMissileDrawable4CD9E0(missile *client.Drawable, hooks magicMissileDrawableHooks4CD9E0) int {
	sparkType := hooks.typeID("Spark")
	_ = hooks.typeID("Puff")
	tailType := hooks.typeID("MagicMissileTailLink")

	prev := image.Pt(int(int32(missile.Field_8)), int(int32(missile.Field_9)))
	dx := int(int32(uint32(missile.PosVec.X) - missile.Field_8))
	dy := int(int32(uint32(missile.PosVec.Y) - missile.Field_9))
	count := hooks.sparkCount()
	if uint32(count) != 0 {
		for i := int32(0); ; {
			pos := image.Pt(
				prev.X+dx*hooks.random(0, 100)/100,
				prev.Y+dy*hooks.random(0, 100)/100,
			)
			spark := hooks.spawn(sparkType, pos)
			if spark != nil {
				effect := spark.UnionEffect()
				effect.Field_108 = uint32(pos.X) << 12
				effect.Field_109 = uint32(pos.Y) << 12
				spark.Field_74_4 = byte(hooks.random(0, 255))
				effect.Field_110 = 0
				effect.Field_112 = hooks.frame() + uint32(hooks.random(3, 10))
				effect.Field_111 = hooks.frame()
				spark.ZVal = 20
				spark.VelZ = int8(hooks.random(0, 6))
				hooks.activate(spark)
			}
			i++
			if i >= count {
				break
			}
		}
	}

	effect := missile.UnionEffect()
	tailDX := uint32(missile.PosVec.X) - effect.Field_108
	tailDY := uint32(missile.PosVec.Y) - effect.Field_109
	if tailDX*tailDX+tailDY*tailDY <= 200 {
		return 1
	}
	tail := hooks.spawn(tailType, image.Pt(
		int(int32(effect.Field_108)),
		int(int32(effect.Field_109)),
	))
	if tail == nil {
		return 1
	}
	tailEffect := tail.UnionEffect()
	tailEffect.Field_108 = uint32(missile.PosVec.X)
	tailEffect.Field_109 = uint32(missile.PosVec.Y)
	hooks.activate(tail)
	effect.Field_108 = uint32(missile.PosVec.X)
	effect.Field_109 = uint32(missile.PosVec.Y)
	hooks.decay(tail, int(hooks.fps()/3))
	return 1
}

func (c *Client) updateMagicMissileDrawable4CD9E0(missile *client.Drawable) int {
	return updateMagicMissileDrawable4CD9E0(missile, magicMissileDrawableHooks4CD9E0{
		sparkCount: func() int32 {
			return memmap.Int32(0x587000, 190108)
		},
		typeID:   c.Things.IndByID,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		fps:      c.srv.TickRate,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
		decay:    c.Objs.TransparentDecay,
	})
}

package opennox

import (
	"image"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/server"
)

type fireballDrawableHooks4CCEA0 struct {
	paused    func() bool
	typeID    func(string) int
	random    func(min, max int) int
	frame     func() uint32
	direction func(image.Point) int
	spawn     func(typ int, pos image.Point) *client.Drawable
	activate  func(*client.Drawable)
}

// updateFireballDrawable4CCE70 replaces all five PE32 fireball updates. The
// legacy helper indexed the native-width drawable as uint32_t[], so PosVec and
// every field after the first widened pointer were read from the wrong bytes.
func updateFireballDrawable4CCE70(ball *client.Drawable, power int, hooks fireballDrawableHooks4CCEA0) int {
	if ball.Field_120 != 0 || hooks.paused() {
		return 1
	}
	sparkType := hooks.typeID("Spark")
	dx := int(int32(uint32(ball.PosVec.X) - ball.Field_8))
	dy := int(int32(uint32(ball.PosVec.Y) - ball.Field_9))
	count := (abs(dx) + abs(dy)) / 7
	for range count {
		fraction := hooks.random(0, 100)
		pos := image.Pt(
			int(int32(ball.Field_8+uint32(dx*fraction/100))),
			int(int32(ball.Field_9+uint32(dy*fraction/100))),
		)
		spark := hooks.spawn(sparkType, pos)
		if spark == nil {
			continue
		}
		effect := spark.UnionEffect()
		effect.Field_108 = uint32(pos.X) << 12
		effect.Field_109 = uint32(pos.Y) << 12
		spark.Field_74_4 = byte(hooks.direction(image.Pt(-dx, -dy)) + hooks.random(-25, 25))
		effect.Field_110 = uint32(power * hooks.random(100, 300))
		effect.Field_112 = hooks.frame() + uint32(hooks.random(30, 45))
		effect.Field_111 = hooks.frame()
		spark.ZVal = 28
		spark.ZVal2 = 0
		spark.VelZ = int8(hooks.random(-2, 4))
		hooks.activate(spark)
	}
	return 1
}

func (c *Client) updateFireballDrawable4CCE70(ball *client.Drawable, power int) int {
	return updateFireballDrawable4CCE70(ball, power, fireballDrawableHooks4CCEA0{
		paused: nox_xxx_checkGameFlagPause_413A50,
		typeID: c.Things.IndByID,
		random: c.srv.Rand.Other.Int,
		frame:  c.srv.Frame,
		direction: func(v image.Point) int {
			return int(server.DirFromVec(types.Pointf{X: float32(v.X), Y: float32(v.Y)}))
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

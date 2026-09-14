package opennox

import (
	"image"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

var (
	charmOrbBright4B6B80    = noxcolor.RGB5551Color(150, 255, 150)
	charmOrbDim4B6B80       = noxcolor.RGB5551Color(0, 220, 0)
	deathBallSparkDim4B6880 = noxcolor.RGB5551Color(100, 255, 50)
)

// callDrawableDraw4B6B80 keeps the two visual effects spawned by Force of
// Nature out of PE32 C draw functions. Other draw functions retain their
// existing dispatch until they are ported separately.
func (c *Client) callDrawableDraw4B6B80(dr *client.Drawable, vp *noxrender.Viewport) int {
	if dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_draw() &&
		int(dr.TypeIDVal) == c.Things.IndByID("CharmOrb") {
		return c.drawCharmOrb4B6B80(dr, vp)
	}
	if dr.DrawFuncPtr == legacy.Get_nox_thing_death_ball_spark_draw() {
		return c.drawDeathBallSpark4B6970(dr, vp)
	}
	return legacy.CallDrawFunc(dr, vp)
}

func charmOrbFields4B6B80(dr *client.Drawable) (radius, tick, countdown byte) {
	data := dr.UnionEffect().Field_111
	return byte(data), byte(data >> 8), byte(data >> 16)
}

func setCharmOrbRadiusAndCountdown4B6B80(dr *client.Drawable, radius, countdown byte) {
	effect := dr.UnionEffect()
	effect.Field_111 = effect.Field_111&0xff00ff00 | uint32(radius) | uint32(countdown)<<16
}

func (c *Client) drawCharmOrb4B6B80(dr *client.Drawable, vp *noxrender.Viewport) int {
	radius, tick, countdown := charmOrbFields4B6B80(dr)
	pos := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -22))
	r := int(radius)
	if pos.X-r >= vp.Screen.Min.X && pos.Y-r >= vp.Screen.Min.Y &&
		pos.X+r < vp.Screen.Max.X && pos.Y+r < vp.Screen.Max.Y {
		c.r.DrawGlow(pos, charmOrbDim4B6B80, r, 5)
		c.r.Data().SetColor2(charmOrbBright4B6B80)
		c.r.DrawPoint(pos, r>>1, charmOrbBright4B6B80)
		old := image.Pt(int(int32(dr.Field_8)), int(int32(dr.Field_9)))
		c.r.DrawLine(pos, pos.Add(old.Sub(dr.PosVec)), charmOrbBright4B6B80)
	}
	if tick == 0 {
		return 1
	}
	countdown--
	if countdown != 0 {
		setCharmOrbRadiusAndCountdown4B6B80(dr, radius, countdown)
		return 1
	}
	countdown = tick
	radius--
	setCharmOrbRadiusAndCountdown4B6B80(dr, radius, countdown)
	if radius == 0 {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	return 1
}

func advanceDeathBallSpark4B6970(dr *client.Drawable, frame uint32) (image.Point, int32, int32) {
	effect := dr.UnionEffect()
	direction := sincosTable16[dr.Field_74_4]
	speed := int32(effect.Field_110)
	effect.Field_108 += uint32(speed * int32(direction.X))
	effect.Field_109 += uint32(speed * int32(direction.Y))
	pos := image.Pt(int(int32(effect.Field_108)>>12), int(int32(effect.Field_109)>>12))
	velocity := dr.VelZ
	z := int16(dr.ZVal) + int16(velocity)
	if z >= 0 {
		dr.ZVal = uint16(z)
		if frame&1 != 0 {
			dr.VelZ = velocity - 1
		}
	} else {
		dr.ZVal = uint16(-z)
		dr.VelZ = int8(-9 * int(velocity) / 10)
		if dr.VelZ < 2 {
			dr.ZVal = 0
			dr.VelZ = 0
		}
	}
	duration := int32(effect.Field_112 - effect.Field_111)
	remaining := int32(effect.Field_112 - frame)
	if remaining == duration {
		remaining--
	}
	return pos, remaining, duration
}

func (c *Client) drawDeathBallSpark4B6970(dr *client.Drawable, vp *noxrender.Viewport) int {
	frame := c.srv.Frame()
	pos, remaining, duration := advanceDeathBallSpark4B6970(dr, frame)
	c.Nox_xxx_updateSpritePosition_49AA90(dr, pos.X, pos.Y)
	if remaining <= 0 || duration <= 0 {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -int(int16(dr.ZVal2))-int(int16(dr.ZVal))))
	if point.X-10 >= vp.Screen.Min.X && point.Y-10 >= vp.Screen.Min.Y &&
		point.X+10 < vp.Screen.Max.X && point.Y+10 < vp.Screen.Max.Y {
		rad := int(4 * remaining / duration)
		c.r.DrawGlow(point, deathBallSparkDim4B6880, 2*rad+1, int(5*remaining/duration))
		c.r.Data().SetColor2(charmOrbBright4B6B80)
		c.r.DrawPoint(point, rad, charmOrbBright4B6B80)
	}
	return 1
}

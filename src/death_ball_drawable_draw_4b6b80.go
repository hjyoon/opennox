package opennox

import (
	"image"
	"math"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

var (
	charmOrbBright4B6B80     = noxcolor.RGB5551Color(150, 255, 150)
	charmOrbDim4B6B80        = noxcolor.RGB5551Color(0, 220, 0)
	healOrbBright4B6B80      = noxcolor.RGB5551Color(255, 255, 0)
	healOrbDim4B6B80         = noxcolor.RGB5551Color(255, 100, 0)
	drainManaOrbBright4B6B80 = noxcolor.RGB5551Color(0, 200, 255)
	drainManaOrbDim4B6B80    = noxcolor.RGB5551Color(0, 0, 255)
	deathBallSparkDim4B6880  = noxcolor.RGB5551Color(100, 255, 50)
	blueSparkBright4B6880    = noxcolor.RGB5551Color(0, 200, 255)
	blueSparkDim4B6880       = noxcolor.RGB5551Color(0, 0, 255)
	cyanSparkBright4B6880    = noxcolor.RGB5551Color(50, 255, 255)
	cyanSparkDim4B6880       = noxcolor.RGB5551Color(0, 200, 200)
	violetSparkBright4B6880  = noxcolor.RGB5551Color(255, 200, 255)
	violetSparkDim4B6880     = noxcolor.RGB5551Color(255, 0, 255)
	manaBombOrbBright4B6B80  = noxcolor.RGB5551Color(255, 255, 255)
	manaBombOrbDim4B6B80     = noxcolor.RGB5551Color(200, 200, 200)
)

// callDrawableDraw4B6B80 keeps migrated glow-orb and spark effects out of
// the PE32 C drawer. Other draw functions retain their existing dispatch.
func (c *Client) callDrawableDraw4B6B80(dr *client.Drawable, vp *noxrender.Viewport) int {
	if dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_draw() ||
		dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_move_draw() {
		switch int(dr.TypeIDVal) {
		case c.Things.IndByID("HealOrb"):
			return c.drawDrainHealOrb4B6B80(dr, vp, healOrbBright4B6B80, healOrbDim4B6B80)
		case c.Things.IndByID("DrainManaOrb"):
			return c.drawDrainHealOrb4B6B80(dr, vp, drainManaOrbBright4B6B80, drainManaOrbDim4B6B80)
		}
	}
	if dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_draw() &&
		int(dr.TypeIDVal) == c.Things.IndByID("CharmOrb") {
		return c.drawCharmOrb4B6B80(dr, vp)
	}
	if dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_draw() &&
		int(dr.TypeIDVal) == c.Things.IndByID("ManaBombOrb") {
		return c.drawManaBombOrb4B6B80(dr, vp)
	}
	if dr.DrawFuncPtr == legacy.Get_nox_thing_death_ball_spark_draw() {
		return c.drawDeathBallSpark4B6970(dr, vp)
	}
	if bright, dim, ok := sparkDrawColors4B6970(dr.DrawFuncPtr); ok {
		return c.drawSpark4B6970(dr, vp, bright, dim)
	}
	return legacy.CallDrawFunc(dr, vp)
}

func sparkDrawColors4B6970(fn unsafe.Pointer) (bright, dim noxcolor.RGBA5551, ok bool) {
	switch fn {
	case legacy.Get_nox_thing_red_spark_draw():
		return healOrbBright4B6B80, healOrbDim4B6B80, true
	case legacy.Get_nox_thing_blue_spark_draw():
		return blueSparkBright4B6880, blueSparkDim4B6880, true
	case legacy.Get_nox_thing_cyan_spark_draw():
		return cyanSparkBright4B6880, cyanSparkDim4B6880, true
	case legacy.Get_nox_thing_green_spark_draw():
		return charmOrbBright4B6B80, charmOrbDim4B6B80, true
	case legacy.Get_nox_thing_yellow_spark_draw():
		return healOrbBright4B6B80, healOrbBright4B6B80, true
	case legacy.Get_nox_thing_violet_spark_draw():
		return violetSparkBright4B6880, violetSparkDim4B6880, true
	case legacy.Get_nox_thing_white_spark_draw():
		return manaBombOrbBright4B6B80, blueSparkBright4B6880, true
	default:
		return 0, 0, false
	}
}

func movingGlowOrbStep4B6B80(dr *client.Drawable) (image.Point, bool) {
	effect := dr.UnionEffect()
	dest := image.Pt(int(uint16(effect.Field_108)), int(uint16(effect.Field_108>>16)))
	dx, dy := dest.X-dr.PosVec.X, dest.Y-dr.PosVec.Y
	distance := int(math.Sqrt(float64(int64(dx)*int64(dx) + int64(dy)*int64(dy))))
	if distance+1 <= 10 {
		return dr.PosVec, true
	}
	speed := int(byte(effect.Field_110 >> 24))
	return image.Pt(dr.PosVec.X+dx*speed/(distance+1), dr.PosVec.Y+dy*speed/(distance+1)), false
}

// HealOrb and DrainManaOrb use the same glow drawer with different colors.
// The moving variant also advances toward the packed source coordinate.
func (c *Client) drawDrainHealOrb4B6B80(dr *client.Drawable, vp *noxrender.Viewport, bright, dim noxcolor.RGBA5551) int {
	if dr.DrawFuncPtr == legacy.Get_nox_thing_glow_orb_move_draw() {
		next, done := movingGlowOrbStep4B6B80(dr)
		if done {
			c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
			return 0
		}
		c.Nox_xxx_updateSpritePosition_49AA90(dr, next.X, next.Y)
	}
	radius, tick, countdown := charmOrbFields4B6B80(dr)
	pos := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -22))
	r := int(radius)
	if pos.X-r >= vp.Screen.Min.X && pos.Y-r >= vp.Screen.Min.Y &&
		pos.X+r < vp.Screen.Max.X && pos.Y+r < vp.Screen.Max.Y {
		c.r.DrawGlow(pos, dim, r, 5)
		c.r.Data().SetColor2(bright)
		c.r.DrawPoint(pos, r>>1, bright)
		old := image.Pt(int(int32(dr.Field_8)), int(int32(dr.Field_9)))
		c.r.DrawLine(pos, pos.Add(old.Sub(dr.PosVec)), bright)
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

// ManaBombOrb uses the white GlowOrb visual. Its C drawer still reads PE32
// field offsets, so draw the native-width drawable through the Go renderer.
func (c *Client) drawManaBombOrb4B6B80(dr *client.Drawable, vp *noxrender.Viewport) int {
	effect := dr.UnionEffect()
	radius := byte(effect.Field_111)
	pos := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -22))
	r := int(radius)
	if pos.X-r >= vp.Screen.Min.X && pos.Y-r >= vp.Screen.Min.Y &&
		pos.X+r < vp.Screen.Max.X && pos.Y+r < vp.Screen.Max.Y {
		c.r.DrawGlow(pos, manaBombOrbDim4B6B80, r, 5)
		c.r.Data().SetColor2(manaBombOrbBright4B6B80)
		c.r.DrawPoint(pos, r>>1, manaBombOrbBright4B6B80)
		old := image.Pt(int(int32(dr.Field_8)), int(int32(dr.Field_9)))
		c.r.DrawLine(pos, pos.Add(old.Sub(dr.PosVec)), manaBombOrbBright4B6B80)
	}
	// sub_499520 creates these orbs with zero fade rate; sub_4CA720 owns
	// their lifetime and removes them when the orbit completes.
	return 1
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
	return c.drawSpark4B6970(dr, vp, charmOrbBright4B6B80, deathBallSparkDim4B6880)
}

func (c *Client) drawSpark4B6970(dr *client.Drawable, vp *noxrender.Viewport, bright, dim noxcolor.RGBA5551) int {
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
		c.r.DrawGlow(point, dim, 2*rad+1, int(5*remaining/duration))
		c.r.Data().SetColor2(bright)
		c.r.DrawPoint(point, rad, bright)
	}
	return 1
}

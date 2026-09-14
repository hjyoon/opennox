package opennox

import (
	"image"
	"math"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type manaBombDrawableHooks4CCAC0 struct {
	radius   func() int
	typeID   func(string) int
	random   func(int, int) int
	frame    func() uint32
	spawn    func(int, image.Point) *client.Drawable
	activate func(*client.Drawable)
}

func (c *Client) manaBombDrawableHooks4CCAC0() manaBombDrawableHooks4CCAC0 {
	return manaBombDrawableHooks4CCAC0{
		radius:   func() int { return int(c.srv.Balance.Float("ManaBombOutRadius")) },
		typeID:   c.Things.IndByID,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	}
}

// updateManaBombCharge4CCAC0 replaces the PE32 client update, which stored
// every spawned drawable pointer in an int before initializing its effect.
func updateManaBombCharge4CCAC0(dr *client.Drawable, hooks manaBombDrawableHooks4CCAC0) int {
	rad := hooks.radius()
	sparkType := hooks.typeID("VioletSpark")
	for range 20 {
		distance := rad/4 + hooks.random(0, rad)
		if distance > rad {
			distance = rad
		}
		angle := uint8(hooks.random(0, 255))
		dir := sincosTable16[angle]
		pos := image.Pt(dr.PosVec.X+distance*dir.X/16, dr.PosVec.Y+distance*dir.Y/16)
		spark := hooks.spawn(sparkType, pos)
		if spark == nil {
			continue
		}
		effect := spark.UnionEffect()
		effect.Field_108 = uint32(spark.PosVec.X) << 12
		effect.Field_109 = uint32(spark.PosVec.Y) << 12
		spark.Field_74_4 = 0
		effect.Field_110 = 0
		effect.Field_112 = hooks.frame() + uint32(hooks.random(10, 30))
		effect.Field_111 = hooks.frame()
		spark.ZVal = 0
		spark.VelZ = int8(hooks.random(2, 8))
		hooks.activate(spark)
	}

	frame := hooks.frame()
	if frame&1 != 0 && frame-dr.Field_80 < 10 {
		orbType := hooks.typeID("ManaBombOrb")
		start := image.Pt(int(int16(dr.PosVec.X)), int(int16(dr.PosVec.Y)))
		for angle := int(frame % 51); angle < 256; angle += 51 {
			dir := sincosTable16[angle]
			end := image.Pt(
				int(int16(start.X+(rad/16)*dir.X)),
				int(int16(start.Y+(rad/16)*dir.Y)),
			)
			spawnManaBombOrb4CA720(orbType, start, end, byte(angle), false, hooks)
			spawnManaBombOrb4CA720(orbType, start, end, byte(angle), true, hooks)
		}
	}
	return 1
}

// spawnManaBombOrb4CA720 mirrors sub_499520 without narrowing a drawable to
// a 32-bit integer. The first two effect words pack PE32's four int16 points.
func spawnManaBombOrb4CA720(typ int, start, end image.Point, angle byte, reverse bool, hooks manaBombDrawableHooks4CCAC0) {
	orb := hooks.spawn(typ, end)
	if orb == nil {
		return
	}
	startX, startY := int16(start.X), int16(start.Y)
	endX, endY := int16(end.X), int16(end.Y)
	dx, dy := int(endX)-int(startX), int(endY)-int(startY)
	length := uint16(math.Sqrt(float64(dx*dx + dy*dy)))
	effect := orb.UnionEffect()
	effect.Field_108 = uint32(uint16(startX)) | uint32(uint16(startY))<<16
	effect.Field_109 = uint32(uint16(endX)) | uint32(uint16(endY))<<16
	effect.Field_110 = uint32(length) | uint32(angle)<<16
	if reverse {
		effect.Field_110 |= 1 << 24
	}
	effect.Field_111 = uint32(byte(hooks.random(3, 10)))
	orb.ClientUpdateFuncPtr = legacy.Get_sub_4CA720()
	orb.Field_127 = uint32(uint16(angle))
	hooks.activate(orb)
}

func manaBombOrbStep4CA720(dr *client.Drawable, frame uint32) (image.Point, bool) {
	effect := dr.UnionEffect()
	age := int32(frame - dr.AnimStart)
	start := image.Pt(int(uint16(effect.Field_108)), int(uint16(effect.Field_108>>16)))
	if age >= 60 ||
		(int(math.Abs(float64(dr.PosVec.X-start.X))) < 10 &&
			int(math.Abs(float64(dr.PosVec.Y-start.Y))) < 10) {
		return image.Point{}, true
	}
	distance := int(int16(effect.Field_110))
	remaining := distance - int(age)*distance/60
	turn := (int(age) << 8) / 120
	angle := int(byte(effect.Field_110 >> 16))
	if byte(effect.Field_110>>24) != 0 {
		angle += turn
	} else {
		angle -= turn
	}
	dir := sincosTable16[uint8(angle)]
	return image.Pt(
		start.X+remaining*dir.X/16,
		start.Y+remaining*dir.Y/16,
	), false
}

func (c *Client) updateManaBombOrb4CA720(dr *client.Drawable) int {
	next, done := manaBombOrbStep4CA720(dr, c.srv.Frame())
	if done {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	c.Nox_xxx_updateSpritePosition_49AA90(dr, next.X, next.Y)
	dr.Field_8, dr.Field_9 = uint32(next.X), uint32(next.Y)
	return 1
}

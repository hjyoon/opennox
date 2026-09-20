//go:build !server

package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type rainDrawKind4B7310 uint8

const (
	rainDrawUnknown4B7310 rainDrawKind4B7310 = iota
	rainDrawBlueRain4B7810
	rainDrawLevelUp4B7700
	rainDrawOblivionUp4B77D0
	rainDrawOrb4B7310
)

func rainDrawKindFor4B7310(fn unsafe.Pointer) rainDrawKind4B7310 {
	switch fn {
	case legacy.Get_nox_thing_blue_rain_draw():
		return rainDrawBlueRain4B7810
	case legacy.Get_nox_thing_levelup_draw():
		return rainDrawLevelUp4B7700
	case legacy.Get_nox_thing_oblivion_up_draw():
		return rainDrawOblivionUp4B77D0
	case legacy.Get_nox_thing_rain_orb_draw():
		return rainDrawOrb4B7310
	default:
		return rainDrawUnknown4B7310
	}
}

type rainSpawnHooks4B7310 struct {
	typeID   func(string) int
	spawn    func(int, image.Point) *client.Drawable
	random   func(int, int) int
	frame    func() uint32
	activate func(*client.Drawable)
}

func spawnBlueRain4B7810(source *client.Drawable, vp *noxrender.Viewport, disabled bool, hooks rainSpawnHooks4B7310) int {
	if disabled {
		return 1
	}
	typeID := hooks.typeID("BlueRainSpark")
	for range 2 {
		pos := source.PosVec.Add(image.Pt(hooks.random(-10, 10), hooks.random(-10, 10)))
		spark := hooks.spawn(typeID, pos)
		if spark == nil {
			continue
		}
		effect := spark.UnionEffect()
		effect.Field_108 = uint32(pos.X) << 12
		effect.Field_109 = uint32(pos.Y) << 12
		spark.Field_74_4 = 0
		effect.Field_110 = 0
		effect.Field_112 = hooks.frame() + uint32(hooks.random(90, 120))
		effect.Field_111 = hooks.frame()
		spark.ZVal2 = 0
		spark.VelZ = -5
		spark.ZVal = uint16(pos.Y - vp.World.Min.Y)
		hooks.activate(spark)
	}
	return 1
}

func spawnFallingRainOrbs4B7740(source *client.Drawable, vp *noxrender.Viewport, typeName string, hooks rainSpawnHooks4B7310) int {
	typeID := hooks.typeID(typeName)
	for range 2 {
		pos := source.PosVec.Add(image.Pt(hooks.random(-15, 15), hooks.random(-15, 15)))
		z := uint16(pos.Y - vp.World.Min.Y)
		velocity := int8(-hooks.random(8, 12))
		orb := hooks.spawn(typeID, pos)
		if orb == nil {
			continue
		}
		orb.ZVal = z
		orb.ZVal2 = 0
		orb.VelZ = velocity
		effect := orb.UnionEffect()
		effect.Field_110 = effect.Field_110&0xff000000 | uint32(z) | uint32(byte(hooks.random(3, 10)))<<16
		effect.Field_108 = uint32(source.PosVec.X)
		effect.Field_109 = uint32(source.PosVec.Y)
		hooks.activate(orb)
	}
	return 1
}

func advanceRainOrbFall4B7310(dr *client.Drawable) (z int16, radius byte, lineDelta int) {
	effect := dr.UnionEffect()
	z = int16(dr.ZVal)
	radius = byte(effect.Field_110 >> 16)
	lineDelta = int(z - int16(uint16(effect.Field_110)))
	effect.Field_110 = effect.Field_110&0xffff0000 | uint32(dr.ZVal)
	dr.ZVal = uint16(z + int16(dr.VelZ))
	return z, radius, lineDelta
}

type rainOrbMoveHooks4B7310 struct {
	typeID    func(string) int
	spawn     func(int, image.Point) *client.Drawable
	random    func(int, int) int
	activate  func(*client.Drawable)
	delete    func(*client.Drawable)
	direction func(types.Pointf) byte
	sinCos    func(byte) (float32, float32)
	round     func(float32) int32
}

func finishRainOrb4B7310(source *client.Drawable, whiteRainOrbType int, hooks rainOrbMoveHooks4B7310) int {
	effect := source.UnionEffect()
	vector := types.Ptf(
		float32(source.PosVec.X-int(int32(effect.Field_108))),
		float32(source.PosVec.Y-int(int32(effect.Field_109))),
	)
	direction := hooks.direction(vector)
	cosine, sine := hooks.sinCos(direction)
	destination := image.Pt(
		int(hooks.round(cosine*150+float32(int32(effect.Field_108)))),
		int(hooks.round(sine*150+float32(int32(effect.Field_109)))),
	)
	moveType := "BlueMoveOrb"
	if int(source.TypeIDVal) == whiteRainOrbType {
		moveType = "WhiteMoveOrb"
	}
	duration := byte(hooks.random(6, 8))
	move := hooks.spawn(hooks.typeID(moveType), image.Pt(int(uint16(source.PosVec.X)), int(uint16(source.PosVec.Y+20))))
	if move != nil {
		moveEffect := move.UnionEffect()
		moveEffect.Field_108 = uint32(uint16(destination.X)) | uint32(uint16(destination.Y))<<16
		moveEffect.Field_110 = moveEffect.Field_110&0x00ffffff | uint32(duration)<<24
		moveEffect.Field_111 = moveEffect.Field_111&0xff000000 |
			uint32(byte(hooks.random(3, 10))) | 1<<8 | 1<<16
		hooks.activate(move)
	}
	hooks.delete(source)
	return 0
}

func (c *Client) rainSpawnHooks4B7310() rainSpawnHooks4B7310 {
	return rainSpawnHooks4B7310{
		typeID:   c.Things.IndByID,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		activate: c.Objs.List34Add,
	}
}

func (c *Client) rainOrbMoveHooks4B7310() rainOrbMoveHooks4B7310 {
	return rainOrbMoveHooks4B7310{
		typeID:   c.Things.IndByID,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		random:   c.srv.Rand.Other.Int,
		activate: c.Objs.List34Add,
		delete:   c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
		direction: func(vector types.Pointf) byte {
			return byte(server.DirFromVec(vector))
		},
		sinCos: server.SinCosDir,
		round:  manaBombCancelFloatToInt48EA70,
	}
}

func (c *Client) drawRainOrb4B7310(dr *client.Drawable, vp *noxrender.Viewport) int {
	whiteRainOrbType := c.Things.IndByID("RainOrbWhite")
	if int16(dr.ZVal) <= 0 {
		return finishRainOrb4B7310(dr, whiteRainOrbType, c.rainOrbMoveHooks4B7310())
	}
	z, radius, lineDelta := advanceRainOrbFall4B7310(dr)
	point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -int(z)))
	glow := blueSparkDim4B6880
	if int(dr.TypeIDVal) == whiteRainOrbType {
		glow = manaBombOrbDim4B6B80
	}
	c.r.DrawGlow(point, glow, int(radius), 5)
	c.r.Data().SetColor2(glow)
	c.r.DrawLine(point, point.Add(image.Pt(0, lineDelta)), glow)
	c.r.Data().SetColor2(manaBombOrbBright4B6B80)
	c.r.DrawPoint(point, int(radius)/3, manaBombOrbBright4B6B80)
	return 1
}

func (c *Client) callRainDrawableDraw4B7310(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch rainDrawKindFor4B7310(dr.DrawFuncPtr) {
	case rainDrawBlueRain4B7810:
		return spawnBlueRain4B7810(dr, vp, noxflags.HasGame(noxflags.GameFlag22), c.rainSpawnHooks4B7310()), true
	case rainDrawLevelUp4B7700:
		return spawnFallingRainOrbs4B7740(dr, vp, "RainOrbWhite", c.rainSpawnHooks4B7310()), true
	case rainDrawOblivionUp4B77D0:
		return spawnFallingRainOrbs4B7740(dr, vp, "RainOrbBlue", c.rainSpawnHooks4B7310()), true
	case rainDrawOrb4B7310:
		return c.drawRainOrb4B7310(dr, vp), true
	default:
		return 0, false
	}
}

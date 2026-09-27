package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
)

const (
	sparkExplosionSparkType48EA70 = iota
	sparkExplosionMediumBoomType48EA70
	sparkExplosionFireBoomType48EA70
)

type sparkExplosionFXState48EA70 struct {
	Pos      image.Point
	Strength byte
}

type sparkExplosionFXHooks48EA70 struct {
	connected func() bool
	types     func() (spark, mediumBoom, fireBoom int)
	random    func(min, max int) int
	frame     func() uint32
	spawn     func(typ int, pos image.Point) *client.Drawable
	activate  func(*client.Drawable)
}

func decodeSparkExplosionFXState48EA70(data []byte) (sparkExplosionFXState48EA70, bool) {
	if len(data) < 6 {
		return sparkExplosionFXState48EA70{}, false
	}
	return sparkExplosionFXState48EA70{
		Pos: image.Pt(
			int(int16(binary.LittleEndian.Uint16(data[1:3]))),
			int(int16(binary.LittleEndian.Uint16(data[3:5]))),
		),
		Strength: data[5],
	}, true
}

func resolveSparkExplosionFXTypesNative48EA70(types *[3]int, lookup func(string) int) (spark, mediumBoom, fireBoom int) {
	if types[sparkExplosionSparkType48EA70] == 0 {
		types[sparkExplosionSparkType48EA70] = lookup("Spark")
		types[sparkExplosionMediumBoomType48EA70] = lookup("MediumFireBoom")
		types[sparkExplosionFireBoomType48EA70] = lookup("FireBoom")
	}
	return types[sparkExplosionSparkType48EA70],
		types[sparkExplosionMediumBoomType48EA70],
		types[sparkExplosionFireBoomType48EA70]
}

func sparkExplosionFXParams48EA70(strength byte) (count, speed, minLife int) {
	v := int(strength)
	return 180*v/255 + 10, 2400*v/255 + 200, 10*v/255 + 5
}

func sparkExplosionBoomType48EA70(strength byte, mediumBoom, fireBoom int) int {
	if strength <= 0xaa {
		return mediumBoom
	}
	return fireBoom
}

func handleSparkExplosionFXNative48EA70(data []byte, hooks sparkExplosionFXHooks48EA70) int {
	state, ok := decodeSparkExplosionFXState48EA70(data)
	if !ok {
		return -1
	}

	// The original client resolves all three types before its connection gate.
	sparkType, mediumBoomType, fireBoomType := hooks.types()
	if !hooks.connected() {
		return 6
	}

	count, speed, minLife := sparkExplosionFXParams48EA70(state.Strength)
	for range count {
		dr := hooks.spawn(sparkType, state.Pos)
		if dr == nil {
			continue
		}
		effect := dr.UnionEffect()
		effect.Field_108 = uint32(dr.PosVec.X) << 12
		effect.Field_109 = uint32(dr.PosVec.Y) << 12
		dr.Field_74_4 = byte(hooks.random(0, 255))
		effect.Field_110 = uint32(hooks.random(1, speed))
		effect.Field_112 = hooks.frame() + uint32(hooks.random(minLife, 96))
		effect.Field_111 = hooks.frame()
		dr.ZVal = uint16(hooks.random(5, 15))
		dr.ZVal2 = 0
		dr.VelZ = int8(hooks.random(0, 8))
		hooks.activate(dr)
	}

	boomType := sparkExplosionBoomType48EA70(state.Strength, mediumBoomType, fireBoomType)
	if dr := hooks.spawn(boomType, state.Pos); dr != nil {
		hooks.activate(dr)
	}
	return 6
}

func (c *Client) handleSparkExplosionFXPacketNative48EA70(data []byte) int {
	return handleSparkExplosionFXNative48EA70(data, sparkExplosionFXHooks48EA70{
		connected: nox_client_isConnected,
		types: func() (int, int, int) {
			return resolveSparkExplosionFXTypesNative48EA70(
				&c.fxSparkExplosionTypes,
				c.Things.IndByID,
			)
		},
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

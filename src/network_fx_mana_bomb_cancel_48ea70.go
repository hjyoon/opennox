package opennox

import (
	"encoding/binary"
	"image"
	"math"

	"github.com/opennox/opennox/v1/client"
)

const manaBombCancelSparkCount48EA70 = 150

type manaBombCancelHooks48EA70 struct {
	connected func() bool
	radius    func() float32
	typeID    func() int
	random    func(min, max int) int
	frame     func() uint32
	spawn     func(typ int, pos image.Point) *client.Drawable
	activate  func(dr *client.Drawable)
}

// manaBombCancelFloatToInt48EA70 models nox_float2int at GAME.EXE 00419A70:
// x87 FISTP under the default round-to-nearest-even mode. Invalid and
// out-of-range conversions produce the signed integer-indefinite value.
func manaBombCancelFloatToInt48EA70(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func resolveManaBombCancelTypeNative48EA70(cache *int, lookup func(string) int) int {
	if *cache == 0 {
		*cache = lookup("CyanSpark")
	}
	return *cache
}

func handleManaBombCancelNative48EA70(data []byte, hooks manaBombCancelHooks48EA70) int {
	if len(data) < 5 {
		return -1
	}
	if !hooks.connected() {
		return 5
	}

	center := image.Pt(
		int(int16(binary.LittleEndian.Uint16(data[1:3]))),
		int(int16(binary.LittleEndian.Uint16(data[3:5]))),
	)
	radius := manaBombCancelFloatToInt48EA70(hooks.radius())
	// The PE32 routine stores the conversion in unsigned v252 before shifting.
	innerRadius := int32(uint32(radius) >> 2)
	typeID := hooks.typeID()
	for range manaBombCancelSparkCount48EA70 {
		distance := innerRadius + int32(hooks.random(0, int(radius)))
		if distance > radius {
			distance = radius
		}
		angle := uint8(hooks.random(0, 255))
		dir := sincosTable16[angle]
		pos := image.Pt(
			int(int32(center.X)+distance*int32(dir.X)/16),
			int(int32(center.Y)+distance*int32(dir.Y)/16),
		)
		dr := hooks.spawn(typeID, pos)
		if dr == nil {
			continue
		}

		effect := dr.UnionEffect()
		effect.Field_108 = uint32(dr.PosVec.X) << 12
		effect.Field_109 = uint32(dr.PosVec.Y) << 12
		dr.Field_74_4 = 0
		effect.Field_110 = 0
		effect.Field_112 = hooks.frame() + uint32(hooks.random(30, 40))
		effect.Field_111 = hooks.frame()
		dr.ZVal = 0
		dr.VelZ = int8(hooks.random(4, 10))
		hooks.activate(dr)
	}
	return 5
}

func (c *Client) handleManaBombCancelPacketNative48EA70(data []byte) int {
	return handleManaBombCancelNative48EA70(data, manaBombCancelHooks48EA70{
		connected: nox_client_isConnected,
		radius: func() float32 {
			return float32(c.srv.Balance.Float("ManaBombOutRadius"))
		},
		typeID: func() int {
			return resolveManaBombCancelTypeNative48EA70(
				&c.fxManaBombCancelType,
				c.Things.IndByID,
			)
		},
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

type pointSparkFXSpec48EA70 struct {
	cache   int
	name    string
	count   int
	speed   int
	minLife int
}

type pointSparkFXHooks48EA70 struct {
	connected func() bool
	typeID    func(pointSparkFXSpec48EA70) int
	random    func(min, max int) int
	frame     func() uint32
	spawn     func(typ int, pos image.Point) *client.Drawable
	activate  func(*client.Drawable)
}

func pointSparkFXSpecForOp48EA70(op netmsg.Op) (pointSparkFXSpec48EA70, bool) {
	switch op {
	case netmsg.MSG_FX_BLUE_SPARKS:
		return pointSparkFXSpec48EA70{0, "BlueSpark", 50, 1000, 30}, true
	case netmsg.MSG_FX_YELLOW_SPARKS:
		return pointSparkFXSpec48EA70{1, "YellowSpark", 25, 500, 25}, true
	case netmsg.MSG_FX_CYAN_SPARKS:
		return pointSparkFXSpec48EA70{2, "CyanSpark", 25, 500, 25}, true
	case netmsg.MSG_FX_VIOLET_SPARKS:
		return pointSparkFXSpec48EA70{3, "VioletSpark", 25, 500, 25}, true
	default:
		return pointSparkFXSpec48EA70{}, false
	}
}

func handlePointSparkFXNative48EA70(op netmsg.Op, data []byte, hooks pointSparkFXHooks48EA70) int {
	spec, ok := pointSparkFXSpecForOp48EA70(op)
	if !ok || len(data) < 5 {
		return -1
	}
	if !hooks.connected() {
		return 5
	}
	pos := image.Pt(
		int(int16(binary.LittleEndian.Uint16(data[1:3]))),
		int(int16(binary.LittleEndian.Uint16(data[3:5]))),
	)
	spawnPointSparkFXNative48EA70(spec, pos, hooks)
	return 5
}

// spawnPointSparkFXNative48EA70 is shared by wire FX packets and local
// effects such as SummonEffect. The legacy helper writes drawable state by
// PE32 byte offsets, so all callers stay on this native-width path.
func spawnPointSparkFXNative48EA70(spec pointSparkFXSpec48EA70, pos image.Point, hooks pointSparkFXHooks48EA70) {
	typ := hooks.typeID(spec)
	for range spec.count {
		dr := hooks.spawn(typ, pos)
		if dr == nil {
			continue
		}
		effect := dr.UnionEffect()
		effect.Field_108 = uint32(dr.PosVec.X) << 12
		effect.Field_109 = uint32(dr.PosVec.Y) << 12
		dr.Field_74_4 = byte(hooks.random(0, 255))
		effect.Field_110 = uint32(hooks.random(1, spec.speed))
		effect.Field_112 = hooks.frame() + uint32(hooks.random(spec.minLife, 64))
		effect.Field_111 = hooks.frame()
		dr.ZVal = 0
		dr.VelZ = int8(hooks.random(2, 10))
		hooks.activate(dr)
	}
}

func (c *Client) handlePointSparkFXPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handlePointSparkFXNative48EA70(op, data, c.pointSparkFXHooksNative48EA70())
}

func (c *Client) spawnPointSparkFXNative48EA70(spec pointSparkFXSpec48EA70, pos image.Point) {
	spawnPointSparkFXNative48EA70(spec, pos, c.pointSparkFXHooksNative48EA70())
}

func (c *Client) pointSparkFXHooksNative48EA70() pointSparkFXHooks48EA70 {
	return pointSparkFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func(spec pointSparkFXSpec48EA70) int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxPointSparkTypes[spec.cache], spec.name, c.Things.IndByID,
			)
		},
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	}
}

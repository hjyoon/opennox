package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

const pointSpriteFXTypeCount48EA70 = 6

type pointSpriteFXSpec48EA70 struct {
	cache int
	name  string
	yoff  int
}

type pointSpriteFXHooks48EA70 struct {
	connected func() bool
	typeID    func(spec pointSpriteFXSpec48EA70) int
	spawn     func(typ int, pos image.Point) *client.Drawable
	activate  func(dr *client.Drawable)
}

func pointSpriteFXSpecForOp48EA70(op netmsg.Op) (pointSpriteFXSpec48EA70, bool) {
	switch op {
	case netmsg.MSG_FX_EXPLOSION:
		return pointSpriteFXSpec48EA70{cache: 0, name: "FireBoom"}, true
	case netmsg.MSG_FX_LESSER_EXPLOSION:
		return pointSpriteFXSpec48EA70{cache: 1, name: "MediumFireBoom"}, true
	case netmsg.MSG_FX_COUNTERSPELL_EXPLOSION:
		return pointSpriteFXSpec48EA70{cache: 2, name: "CounterspellBoom"}, true
	case netmsg.MSG_FX_THIN_EXPLOSION:
		return pointSpriteFXSpec48EA70{cache: 3, name: "ThinFireBoom"}, true
	case netmsg.MSG_FX_TELEPORT:
		return pointSpriteFXSpec48EA70{cache: 4, name: "TeleportPoof", yoff: 2}, true
	case netmsg.MSG_FX_DAMAGE_POOF:
		return pointSpriteFXSpec48EA70{cache: 5, name: "DamagePoof", yoff: 2}, true
	default:
		return pointSpriteFXSpec48EA70{}, false
	}
}

func resolvePointSpriteFXTypeNative48EA70(cache *int, name string, lookup func(string) int) int {
	if *cache == 0 {
		*cache = lookup(name)
	}
	return *cache
}

func handlePointSpriteFXNative48EA70(op netmsg.Op, data []byte, hooks pointSpriteFXHooks48EA70) int {
	spec, ok := pointSpriteFXSpecForOp48EA70(op)
	if !ok || len(data) < 5 {
		return -1
	}
	if !hooks.connected() {
		return 5
	}
	pos := image.Pt(
		int(int16(binary.LittleEndian.Uint16(data[1:3]))),
		int(int16(binary.LittleEndian.Uint16(data[3:5])))+spec.yoff,
	)
	if dr := hooks.spawn(hooks.typeID(spec), pos); dr != nil {
		hooks.activate(dr)
	}
	return 5
}

func (c *Client) handlePointSpriteFXPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handlePointSpriteFXNative48EA70(op, data, pointSpriteFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func(spec pointSpriteFXSpec48EA70) int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxPointSpriteTypes[spec.cache],
				spec.name,
				c.Things.IndByID,
			)
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
)

var sphericalShieldNames48EA70 = [...]string{
	"SphericalShieldNW",
	"SphericalShieldN",
	"SphericalShieldNE",
	"SphericalShieldW",
	"",
	"SphericalShieldE",
	"SphericalShieldSW",
	"SphericalShieldS",
	"SphericalShieldSE",
}

type shieldFXHooks48EA70 struct {
	connected func() bool
	byCode    func(uint16) *client.Drawable
	typeID    func(int, string) int
	eachNear  func(image.Rectangle, func(*client.Drawable))
	spawn     func(int, image.Point) *client.Drawable
}

func handleShieldFXNative48EA70(data []byte, hooks shieldFXHooks48EA70) int {
	if len(data) < 4 {
		return -1
	}
	if !hooks.connected() {
		return 4
	}
	code := binary.LittleEndian.Uint16(data[1:3])
	direction := int(data[3])
	if direction < 0 || direction >= len(sphericalShieldNames48EA70) {
		return 4
	}
	target := hooks.byCode(code)
	if target == nil {
		return 4
	}
	var types [len(sphericalShieldNames48EA70)]int
	for i, name := range sphericalShieldNames48EA70 {
		if name != "" {
			types[i] = hooks.typeID(i, name)
		}
	}
	rect := image.Rect(target.PosVec.X-10, target.PosVec.Y-10, target.PosVec.X+11, target.PosVec.Y+11)
	duplicate := false
	hooks.eachNear(rect, func(dr *client.Drawable) {
		if duplicate || dr == nil || !dr.PosVec.In(rect) {
			return
		}
		for _, typ := range types {
			if typ != 0 && dr.TypeIDVal == uint32(typ) && dr.UnionEffect().Field_108 == uint32(code) {
				duplicate = true
				return
			}
		}
	})
	if duplicate || types[direction] == 0 {
		return 4
	}
	dr := hooks.spawn(types[direction], image.Pt(target.PosVec.X, target.PosVec.Y+3))
	if dr != nil {
		dr.UnionEffect().Field_108 = uint32(code)
	}
	return 4
}

func (c *Client) handleShieldFXPacketNative48EA70(data []byte) int {
	return handleShieldFXNative48EA70(data, shieldFXHooks48EA70{
		connected: nox_client_isConnected,
		byCode:    c.Objs.ByNetCode,
		typeID: func(_ int, name string) int {
			return c.Things.IndByID(name)
		},
		eachNear: c.Objs.EachInRect,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
	})
}

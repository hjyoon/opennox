package opennox

import (
	"encoding/binary"
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const (
	turnUndeadFXPacketSize48EA70    = 5
	turnUndeadFXDrawableCount48EA70 = 43
)

type turnUndeadFXHooks48EA70 struct {
	connected       func() bool
	typeID          func() int
	frame           func() uint32
	directionVector func(byte) (float32, float32)
	spawn           func(int, image.Point) *client.Drawable
	secondaryUpdate unsafe.Pointer
	activate        func(*client.Drawable)
	sightDestroy    func(*client.Drawable)
}

func decodeTurnUndeadFXPosition48EA70(data []byte) (image.Point, bool) {
	if len(data) < turnUndeadFXPacketSize48EA70 {
		return image.Point{}, false
	}
	return image.Pt(
		int(int16(binary.LittleEndian.Uint16(data[1:3]))),
		int(int16(binary.LittleEndian.Uint16(data[3:5]))),
	), true
}

func handleTurnUndeadFXNative48EA70(data []byte, hooks turnUndeadFXHooks48EA70) int {
	pos, ok := decodeTurnUndeadFXPosition48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return turnUndeadFXPacketSize48EA70
	}

	typeID := hooks.typeID()
	for direction := 0; direction < 256; direction += 6 {
		dr := hooks.spawn(typeID, pos)
		if dr == nil {
			continue
		}
		cosine, sine := hooks.directionVector(byte(direction))

		dr.SetTurnUndeadEffect(byte(direction), pos, cosine, sine, hooks.frame(), hooks.secondaryUpdate)
		hooks.activate(dr)
		hooks.sightDestroy(dr)
	}
	return turnUndeadFXPacketSize48EA70
}

func (c *Client) handleTurnUndeadFXPacketNative48EA70(data []byte) int {
	return handleTurnUndeadFXNative48EA70(data, turnUndeadFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxTurnUndeadType, "UndeadKiller", c.Things.IndByID,
			)
		},
		frame:           c.srv.Frame,
		directionVector: server.SinCosDir,
		spawn:           c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		secondaryUpdate: legacy.Get_nox_xxx_sprite_4CA540(),
		activate:        c.Objs.List5Add,
		sightDestroy:    c.Objs.List6Add,
	})
}

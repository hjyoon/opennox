package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

type deltaZFXState48EA70 struct {
	Code     uint16
	Height   byte
	Velocity int8
	Target   byte
}

type deltaZFXHooks48EA70 struct {
	connected func() bool
	byCode    func(uint16) *client.Drawable
	frame     func() uint32
}

func decodeDeltaZFXState48EA70(data []byte) (deltaZFXState48EA70, bool) {
	if len(data) < 6 {
		return deltaZFXState48EA70{}, false
	}
	return deltaZFXState48EA70{
		Code:     binary.LittleEndian.Uint16(data[1:3]),
		Height:   data[3],
		Velocity: int8(data[4]),
		Target:   data[5],
	}, true
}

func handleDeltaZFXNative48EA70(data []byte, hooks deltaZFXHooks48EA70) int {
	state, ok := decodeDeltaZFXState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 6
	}
	dr := hooks.byCode(state.Code)
	if dr == nil {
		return 6
	}
	effect := dr.UnionEffect()
	effect.Field_108 = hooks.frame()
	effect.Field_109 = math.Float32bits(float32(state.Height))
	effect.Field_110 = math.Float32bits(float32(state.Velocity))
	effect.Field_111 = math.Float32bits(float32(state.Target))
	return 6
}

type arrowTrapFXState48EA70 struct {
	Position image.Point
	Variant  byte
}

type arrowTrapFXHooks48EA70 struct {
	connected func() bool
	typeIDs   func() [2]int
	spawn     func(int, image.Point) *client.Drawable
	activate  func(*client.Drawable)
}

func decodeArrowTrapFXState48EA70(data []byte) (arrowTrapFXState48EA70, bool) {
	if len(data) < 6 {
		return arrowTrapFXState48EA70{}, false
	}
	return arrowTrapFXState48EA70{
		Position: image.Pt(
			int(int16(binary.LittleEndian.Uint16(data[1:3]))),
			int(int16(binary.LittleEndian.Uint16(data[3:5]))),
		),
		Variant: data[5],
	}, true
}

func handleArrowTrapFXNative48EA70(data []byte, hooks arrowTrapFXHooks48EA70) int {
	state, ok := decodeArrowTrapFXState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 6
	}
	types := hooks.typeIDs()
	index := 1
	pos := state.Position.Add(image.Pt(-3, 0))
	if state.Variant == 1 {
		index = 0
		pos = state.Position.Add(image.Pt(15, 0))
	}
	if dr := hooks.spawn(types[index], pos); dr != nil {
		hooks.activate(dr)
	}
	return 6
}

type vampirismFXState48EA70 struct {
	From   image.Point
	To     image.Point
	Amount uint16
}

type vampirismFXHooks48EA70 struct {
	connected func() bool
	typeID    func() int
	random    func(int, int) int
	spawn     func(int, image.Point) *client.Drawable
	activate  func(*client.Drawable)
}

func decodeVampirismFXState48EA70(data []byte) (vampirismFXState48EA70, bool) {
	if len(data) < 11 {
		return vampirismFXState48EA70{}, false
	}
	return vampirismFXState48EA70{
		From: image.Pt(
			int(binary.LittleEndian.Uint16(data[1:3])),
			int(binary.LittleEndian.Uint16(data[3:5])),
		),
		To: image.Pt(
			int(binary.LittleEndian.Uint16(data[5:7])),
			int(binary.LittleEndian.Uint16(data[7:9])),
		),
		Amount: binary.LittleEndian.Uint16(data[9:11]),
	}, true
}

func vampirismFXOrbCount48EA70(amount uint16) int {
	count := int(amount >> 2)
	if count > 7 {
		count = 7
	}
	return count + 1
}

func setOrbFXPayloadNative48EA70(dr *client.Drawable, destination image.Point, radius, fade, mode byte) {
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 15)
	binary.LittleEndian.PutUint16(payload[0:2], uint16(destination.X))
	binary.LittleEndian.PutUint16(payload[2:4], uint16(destination.Y))
	payload[11] = radius
	payload[12] = fade
	payload[13] = mode
	payload[14] = mode
}

func handleVampirismFXNative48EA70(data []byte, hooks vampirismFXHooks48EA70) int {
	state, ok := decodeVampirismFXState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 11
	}
	typ := hooks.typeID()
	for range vampirismFXOrbCount48EA70(state.Amount) {
		radius := byte(hooks.random(6, 12))
		dy := hooks.random(-20, 20)
		dx := hooks.random(-20, 20)
		dr := hooks.spawn(typ, state.To.Add(image.Pt(dx, dy)))
		if dr == nil {
			continue
		}
		setOrbFXPayloadNative48EA70(dr, state.From, radius, byte(hooks.random(3, 10)), 0)
		hooks.activate(dr)
	}
	return 11
}

func (c *Client) handleDeltaZFXPacketNative48EA70(data []byte) int {
	return handleDeltaZFXNative48EA70(data, deltaZFXHooks48EA70{
		connected: nox_client_isConnected,
		byCode:    c.Objs.ByNetCode,
		frame:     c.srv.Frame,
	})
}

func (c *Client) handleArrowTrapFXPacketNative48EA70(data []byte) int {
	return handleArrowTrapFXNative48EA70(data, arrowTrapFXHooks48EA70{
		connected: nox_client_isConnected,
		typeIDs: func() [2]int {
			if c.fxArrowTrapTypes[0] == 0 {
				c.fxArrowTrapTypes[0] = c.Things.IndByID("ArrowTrap1Smoke")
				c.fxArrowTrapTypes[1] = c.Things.IndByID("ArrowTrap2Smoke")
			}
			return c.fxArrowTrapTypes
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

func (c *Client) handleVampirismFXPacketNative48EA70(data []byte) int {
	return handleVampirismFXNative48EA70(data, vampirismFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxVampirismType, "HealOrb", c.Things.IndByID,
			)
		},
		random:   c.srv.Rand.Other.Int,
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate: c.Objs.List34Add,
	})
}

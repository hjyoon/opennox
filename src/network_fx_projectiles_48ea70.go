package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
)

type sentryRayFXState48EA70 struct {
	From image.Point
	To   image.Point
}

type sentryRayFXHooks48EA70 struct {
	connected func() bool
	queueRay  func(from, to image.Point)
	random    func(min, max int) int
	playerPos func() (image.Point, bool)
	playSound func(sound.ID, int)
	paused    func() bool
	typeID    func() int
	spawn     func(typ int, pos image.Point) *client.Drawable
	frame     func() uint32
	activate  func(*client.Drawable)
}

func decodeSentryRayFXState48EA70(data []byte) (sentryRayFXState48EA70, bool) {
	if len(data) < 9 {
		return sentryRayFXState48EA70{}, false
	}
	return sentryRayFXState48EA70{
		From: image.Pt(
			int(binary.LittleEndian.Uint16(data[1:3])),
			int(binary.LittleEndian.Uint16(data[3:5])),
		),
		To: image.Pt(
			int(binary.LittleEndian.Uint16(data[5:7])),
			int(binary.LittleEndian.Uint16(data[7:9])),
		),
	}, true
}

func handleSentryRayFXNative48EA70(data []byte, hooks sentryRayFXHooks48EA70) int {
	state, ok := decodeSentryRayFXState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 9
	}

	hooks.queueRay(state.From, state.To)
	if hooks.random(0, 100) < 25 {
		if player, ok := hooks.playerPos(); ok {
			delta := state.To.Sub(player)
			distance := int(math.Sqrt(float64(int64(delta.X)*int64(delta.X) + int64(delta.Y)*int64(delta.Y))))
			if distance < 600 {
				hooks.playSound(sound.SoundSentryRayHitWall, 100*(600-distance)/600)
			}
		}
	}
	if hooks.paused() {
		return 9
	}

	delta := state.To.Sub(state.From)
	distance := int(math.Sqrt(float64(int64(delta.X)*int64(delta.X) + int64(delta.Y)*int64(delta.Y))))
	if distance == 0 {
		distance = 1
	}
	pos := image.Pt(
		state.To.X-4*delta.X/distance,
		state.To.Y-4*delta.Y/distance,
	)
	dr := hooks.spawn(hooks.typeID(), pos)
	if dr == nil {
		return 9
	}
	effect := dr.UnionEffect()
	effect.Field_108 = uint32(pos.X) << 12
	effect.Field_109 = uint32(pos.Y) << 12
	dr.Field_74_4 = byte(hooks.random(0, 255))
	effect.Field_110 = uint32(hooks.random(1, 1500))
	effect.Field_112 = hooks.frame() + uint32(hooks.random(5, 20))
	effect.Field_111 = hooks.frame()
	dr.ZVal = 22
	dr.VelZ = int8(hooks.random(-4, 4))
	hooks.activate(dr)
	return 9
}

type ricochetFXHooks48EA70 struct {
	connected func() bool
	typeID    func() int
	spawn     func(typ int, pos image.Point) *client.Drawable
	random    func(min, max int) int
	frame     func() uint32
	activate  func(*client.Drawable)
}

func decodeRicochetFXPosition48EA70(data []byte) (image.Point, bool) {
	if len(data) < 5 {
		return image.Point{}, false
	}
	return image.Pt(
		int(int16(binary.LittleEndian.Uint16(data[1:3]))),
		int(int16(binary.LittleEndian.Uint16(data[3:5]))),
	), true
}

func handleRicochetFXNative48EA70(data []byte, hooks ricochetFXHooks48EA70) int {
	pos, ok := decodeRicochetFXPosition48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 5
	}
	typ := hooks.typeID()
	for range 5 {
		dr := hooks.spawn(typ, pos)
		if dr == nil {
			continue
		}
		effect := dr.UnionEffect()
		effect.Field_108 = uint32(dr.PosVec.X) << 12
		effect.Field_109 = uint32(dr.PosVec.Y) << 12
		dr.Field_74_4 = byte(hooks.random(0, 255))
		effect.Field_110 = uint32(hooks.random(1333, 4000))
		effect.Field_112 = hooks.frame() + uint32(hooks.random(5, 20))
		effect.Field_111 = hooks.frame()
		dr.ZVal = 20
		dr.VelZ = int8(hooks.random(-5, 5))
		hooks.activate(dr)
	}
	return 5
}

type greenBoltFXState48EA70 struct {
	From     image.Point
	To       image.Point
	Duration uint16
}

type greenBoltFXHooks48EA70 struct {
	connected func() bool
	typeID    func() int
	spawn     func(typ int, pos image.Point) *client.Drawable
}

func decodeGreenBoltFXState48EA70(data []byte) (greenBoltFXState48EA70, bool) {
	if len(data) < 11 {
		return greenBoltFXState48EA70{}, false
	}
	return greenBoltFXState48EA70{
		From: image.Pt(
			int(binary.LittleEndian.Uint16(data[1:3])),
			int(binary.LittleEndian.Uint16(data[3:5])),
		),
		To: image.Pt(
			int(binary.LittleEndian.Uint16(data[5:7])),
			int(binary.LittleEndian.Uint16(data[7:9])),
		),
		Duration: binary.LittleEndian.Uint16(data[9:11]),
	}, true
}

func greenBoltFXMidpoint48EA70(state greenBoltFXState48EA70) image.Point {
	return image.Pt(
		state.From.X+(state.To.X-state.From.X)/2,
		state.From.Y+(state.To.Y-state.From.Y)/2,
	)
}

func setGreenBoltFXPayloadNative48EA70(dr *client.Drawable, state greenBoltFXState48EA70) {
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 13)
	payload[0] = 0
	binary.LittleEndian.PutUint32(payload[1:5], uint32(state.Duration))
	binary.LittleEndian.PutUint16(payload[5:7], uint16(state.From.X))
	binary.LittleEndian.PutUint16(payload[7:9], uint16(state.From.Y))
	binary.LittleEndian.PutUint16(payload[9:11], uint16(state.To.X))
	binary.LittleEndian.PutUint16(payload[11:13], uint16(state.To.Y))
}

func handleGreenBoltFXNative48EA70(data []byte, hooks greenBoltFXHooks48EA70) int {
	state, ok := decodeGreenBoltFXState48EA70(data)
	if !ok {
		return -1
	}
	// Vanilla resolves GreenZap before checking the connection state.
	typ := hooks.typeID()
	if !hooks.connected() {
		return 11
	}
	if dr := hooks.spawn(typ, greenBoltFXMidpoint48EA70(state)); dr != nil {
		setGreenBoltFXPayloadNative48EA70(dr, state)
	}
	return 11
}

func (c *Client) handleSentryRayFXPacketNative48EA70(data []byte) int {
	return handleSentryRayFXNative48EA70(data, sentryRayFXHooks48EA70{
		connected: nox_client_isConnected,
		queueRay:  legacy.AddSentryRay4C5020,
		random:    c.srv.Rand.Other.Int,
		playerPos: func() (image.Point, bool) {
			player := c.ClientPlayerUnit()
			if player == nil {
				return image.Point{}, false
			}
			return player.PosVec, true
		},
		playSound: clientPlaySoundSpecial,
		paused:    nox_xxx_checkGameFlagPause_413A50,
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxPointSparkTypes[3], "VioletSpark", c.Things.IndByID,
			)
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		frame:    c.srv.Frame,
		activate: c.Objs.List34Add,
	})
}

func (c *Client) handleRicochetFXPacketNative48EA70(data []byte) int {
	return handleRicochetFXNative48EA70(data, ricochetFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxPointSparkTypes[0], "BlueSpark", c.Things.IndByID,
			)
		},
		spawn:    c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		random:   c.srv.Rand.Other.Int,
		frame:    c.srv.Frame,
		activate: c.Objs.List34Add,
	})
}

func (c *Client) handleGreenBoltFXPacketNative48EA70(data []byte) int {
	return handleGreenBoltFXNative48EA70(data, greenBoltFXHooks48EA70{
		connected: nox_client_isConnected,
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxGreenBoltType, "GreenZap", c.Things.IndByID,
			)
		},
		spawn: c.Nox_xxx_spriteLoadAdd_45A360_drawable,
	})
}

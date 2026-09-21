package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

const clientPredictLinearPacketSize48EA70 = 14

type clientPredictLinearState48EA70 struct {
	Code      uint16
	TypeID    uint16
	Pos       image.Point
	Field127  uint16
	Damping   int8
	VelocityX int8
	VelocityY int8
}

type clientPredictLinearHooks48EA70 struct {
	connected       func() bool
	create          func(typeID int, code uint16, x, y int) *client.Drawable
	frame           func() uint32
	secondaryUpdate unsafe.Pointer
	activate        func(*client.Drawable)
}

func decodeClientPredictLinearState48EA70(data []byte) (clientPredictLinearState48EA70, bool) {
	if len(data) < clientPredictLinearPacketSize48EA70 {
		return clientPredictLinearState48EA70{}, false
	}
	return clientPredictLinearState48EA70{
		Code:   binary.LittleEndian.Uint16(data[1:3]),
		TypeID: binary.LittleEndian.Uint16(data[3:5]),
		Pos: image.Pt(
			int(binary.LittleEndian.Uint16(data[5:7])),
			int(binary.LittleEndian.Uint16(data[7:9])),
		),
		Field127:  binary.LittleEndian.Uint16(data[9:11]),
		Damping:   int8(data[11]),
		VelocityX: int8(data[12]),
		VelocityY: int8(data[13]),
	}, true
}

func handleClientPredictLinearNative48EA70(data []byte, hooks clientPredictLinearHooks48EA70) int {
	state, ok := decodeClientPredictLinearState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return clientPredictLinearPacketSize48EA70
	}

	// GAME.EXE explicitly clears the static-object namespace bit before
	// creating this short-lived predicted drawable.
	dr := hooks.create(
		int(state.TypeID),
		nox_xxx_netClearHighBit_578B30(state.Code),
		state.Pos.X,
		state.Pos.Y,
	)
	if dr == nil {
		return clientPredictLinearPacketSize48EA70
	}

	// These are named native fields corresponding to PE32 fields 127 and
	// 117..119. Raw offsets 508 and 468..476 do not describe a widened
	// Drawable and can overwrite pointers on 64-bit hosts.
	dr.Field_127 = dr.Field_127&0xffff0000 | uint32(state.Field127)
	dr.Field_117 = math.Float32bits(float32(state.VelocityX) * 0.0625)
	dr.Field_118 = math.Float32bits(float32(state.VelocityY) * 0.0625)
	dr.Field_119 = math.Float32bits(float32(state.Damping) * 0.0625)
	dr.AnimStart = hooks.frame()
	dr.Field_81 = uint32(state.Pos.X)
	dr.Field_82 = uint32(state.Pos.Y)
	dr.Field_115 = hooks.secondaryUpdate
	hooks.activate(dr)
	return clientPredictLinearPacketSize48EA70
}

func (c *Client) handleClientPredictLinearPacketNative48EA70(data []byte) int {
	return handleClientPredictLinearNative48EA70(data, clientPredictLinearHooks48EA70{
		connected:       nox_client_isConnected,
		create:          c.Nox_xxx_spriteCreate_48E970,
		frame:           c.Server.Frame,
		secondaryUpdate: legacy.Get_nox_xxx_sprite_4CA540(),
		activate:        c.Objs.List5Add,
	})
}

type clientPredictLinearUpdateHooks4CA540 struct {
	frame   func() uint32
	move    func(*client.Drawable, int, int)
	visible func(int, int) int
	remove  func(*client.Drawable)
}

// clientPredictLinearFloatToInt4CA540 models nox_float2int's x87 FISTP
// conversion under the default round-to-nearest-even mode.
func clientPredictLinearFloatToInt4CA540(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

// clientPredictLinearPosition4CA540 reconstructs the predicted position from
// the packet's initial point on every frame, matching GAME.EXE 004CA540.
func clientPredictLinearPosition4CA540(dr *client.Drawable, frame uint32) image.Point {
	damping := float64(math.Float32frombits(dr.Field_119))
	velocityX := float64(math.Float32frombits(dr.Field_117))
	velocityY := float64(math.Float32frombits(dr.Field_118))
	var displacementX, displacementY float64

	steps := frame - dr.AnimStart + 1
	for ; steps != 0; steps-- {
		// The original stores the Y deceleration in a float temporary before
		// adding it back to the x87 value. Preserve that single-precision step.
		deltaY := float32(-(velocityY * damping))
		velocityX -= velocityX * damping
		velocityY += float64(deltaY)
		displacementX += velocityX
		displacementY += velocityY
	}

	xValue := float32(float64(int32(dr.Field_81)) + displacementX)
	// GAME.EXE rounds the accumulated Y displacement to float before adding
	// the initial coordinate.
	yValue := float32(float64(int32(dr.Field_82)) + float64(float32(displacementY)))
	return image.Pt(
		int(clientPredictLinearFloatToInt4CA540(xValue)),
		int(clientPredictLinearFloatToInt4CA540(yValue)),
	)
}

func updateClientPredictLinearNative4CA540(vp *noxrender.Viewport, dr *client.Drawable, hooks clientPredictLinearUpdateHooks4CA540) int {
	pos := clientPredictLinearPosition4CA540(dr, hooks.frame())
	if pos.X > 0 && pos.Y > 0 && pos.X < 5888 && pos.Y < 5888 {
		hooks.move(dr, pos.X, pos.Y)
		screen := dr.PosVec
		if vp != nil {
			screen = vp.ToScreenPos(dr.PosVec)
		}
		if hooks.visible(screen.X, screen.Y) != 0 {
			return 1
		}
	}
	hooks.remove(dr)
	return 0
}

func (c *Client) updateClientPredictLinear4CA540(vp *noxrender.Viewport, dr *client.Drawable) int {
	return updateClientPredictLinearNative4CA540(vp, dr, clientPredictLinearUpdateHooks4CA540{
		frame:   c.Server.Frame,
		move:    c.Nox_xxx_updateSpritePosition_49AA90,
		visible: c.Sight.Sub_4992B0,
		remove:  c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
	})
}

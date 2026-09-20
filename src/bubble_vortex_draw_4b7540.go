//go:build !server

package opennox

import (
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

type bubbleVortexDrawKind4B7540 uint8

const (
	bubbleVortexDrawUnknown4B7540 bubbleVortexDrawKind4B7540 = iota
	bubbleVortexDrawBubble4B7540
	bubbleVortexDrawVortex4B9F50
)

func bubbleVortexDrawKindFor4B7540(fn unsafe.Pointer) bubbleVortexDrawKind4B7540 {
	switch fn {
	case legacy.Get_nox_thing_bubble_draw():
		return bubbleVortexDrawBubble4B7540
	case legacy.Get_nox_thing_vortex_draw():
		return bubbleVortexDrawVortex4B9F50
	default:
		return bubbleVortexDrawUnknown4B7540
	}
}

func (c *Client) callBubbleVortexDraw4B7540(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch bubbleVortexDrawKindFor4B7540(dr.DrawFuncPtr) {
	case bubbleVortexDrawBubble4B7540:
		return c.drawBubble4B7540(dr, vp), true
	case bubbleVortexDrawVortex4B9F50:
		return c.drawVortex4B9F50(dr, vp), true
	default:
		return 0, false
	}
}

func packedEffectByte4B7540(value uint32, shift uint) byte {
	return byte(value >> shift)
}

func setPackedEffectByte4B7540(value uint32, shift uint, next byte) uint32 {
	mask := uint32(0xff) << shift
	return value&^mask | uint32(next)<<shift
}

type bubbleDrawVisual4B7540 struct {
	point       image.Point
	baseRadius  byte
	glowColor   uint32
	centerColor uint32
}

func bubbleDrawVisualFor4B7540(dr *client.Drawable, vp *noxrender.Viewport) bubbleDrawVisual4B7540 {
	effect := dr.UnionEffect()
	return bubbleDrawVisual4B7540{
		point:       vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -int(int16(dr.ZVal)))),
		baseRadius:  packedEffectByte4B7540(effect.Field_110, 0),
		glowColor:   effect.Field_108,
		centerColor: effect.Field_109,
	}
}

// prepareBubbleDraw4B7540 performs the checks that preceded the original
// draw calls. A deadline transition starts a fresh one-second transparent
// decay before the radius-zero removal check.
func prepareBubbleDraw4B7540(dr *client.Drawable, frame uint32) (startDecay, remove bool) {
	effect := dr.UnionEffect()
	radius := packedEffectByte4B7540(effect.Field_110, 0)
	phase := packedEffectByte4B7540(effect.Field_110, 8)
	if phase == 3 {
		return false, radius == 0
	}
	if dr.Deadline != 0 && dr.Deadline <= frame {
		effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 8, 3)
		effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 16, 4)
		effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 24, 4)
		return true, radius == 0
	}
	return false, false
}

// advanceBubbleDraw4B7540 applies the post-draw animation state. It returns
// true when the caller must remove the drawable after the frame was drawn.
func advanceBubbleDraw4B7540(dr *client.Drawable, frame uint32) (remove bool) {
	effect := dr.UnionEffect()
	if byte(frame)&3 != 0 {
		step := int8(packedEffectByte4B7540(effect.Field_111, 16))
		dr.ZVal = uint16(int32(int16(dr.ZVal)) + int32(step))
	}

	timer := packedEffectByte4B7540(effect.Field_110, 16)
	if timer != 0 {
		timer--
		effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 16, timer)
		if timer == 0 {
			phase := packedEffectByte4B7540(effect.Field_110, 8)
			radius := packedEffectByte4B7540(effect.Field_110, 0)
			switch phase {
			case 1:
				radius++
				effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 0, radius)
				if radius >= 5 {
					effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 8, 2)
				}
			case 2:
				radius--
				effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 0, radius)
				if radius == 0 {
					effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 8, 1)
				}
			default:
				radius--
				effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 0, radius)
				if radius == 0 {
					return true
				}
			}
			reset := packedEffectByte4B7540(effect.Field_110, 24)
			effect.Field_110 = setPackedEffectByte4B7540(effect.Field_110, 16, reset)
		}
	}

	zTimer := packedEffectByte4B7540(effect.Field_111, 8)
	if zTimer != 0 {
		zTimer--
		effect.Field_111 = setPackedEffectByte4B7540(effect.Field_111, 8, zTimer)
		if zTimer == 0 {
			reset := packedEffectByte4B7540(effect.Field_111, 0)
			step := packedEffectByte4B7540(effect.Field_111, 16)
			effect.Field_111 = setPackedEffectByte4B7540(effect.Field_111, 16, byte(0-step))
			effect.Field_111 = setPackedEffectByte4B7540(effect.Field_111, 8, reset)
		}
	}
	return int16(dr.ZVal) < 0
}

// BubbleDraw interpreted the native Drawable as a PE32 byte array. Use the
// numeric effect view, whose packed 20-byte prefix is stable on 32- and
// 64-bit builds, and keep all object access at native pointer width.
func (c *Client) drawBubble4B7540(dr *client.Drawable, vp *noxrender.Viewport) int {
	startDecay, remove := prepareBubbleDraw4B7540(dr, c.srv.Frame())
	if startDecay {
		c.Objs.TransparentDecay(dr, int(c.srv.TickRate()))
	}
	if remove {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}

	visual := bubbleDrawVisualFor4B7540(dr, vp)
	glowColor := noxcolor.RGBA5551(visual.glowColor)
	centerColor := noxcolor.RGBA5551(visual.centerColor)
	c.r.DrawGlow(visual.point, glowColor, int(visual.baseRadius), int(visual.baseRadius)+3)
	c.r.DrawPoint(visual.point, int(visual.baseRadius>>1), centerColor)
	if advanceBubbleDraw4B7540(dr, c.srv.Frame()) {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	return 1
}

type vortexDrawState4B9F50 struct {
	point       image.Point
	tail        image.Point
	pointColor  uint32
	tailColor   uint32
	glowSize    int
	pointRadius int
	inside      bool
	alive       bool
	nextAngle   byte
	nextRadius  byte
	nextZ       uint16
}

var vortexGray4B9F50 = nox_color_rgb_4344A0(170, 170, 170)

func vortexRadiusForZ4B9F50(z int16) int {
	scaled := float32(float64(z) * 0.0024999999 * 50.0)
	return 50 - int(manaBombCancelFloatToInt48EA70(scaled))
}

func vortexDrawStateFor4B9F50(dr *client.Drawable, vp *noxrender.Viewport) vortexDrawState4B9F50 {
	effect := dr.UnionEffect()
	motion := effect.Field_112
	angle := packedEffectByte4B7540(motion, 0)
	spin := int8(packedEffectByte4B7540(motion, 8))
	radius := int(packedEffectByte4B7540(motion, 16))
	rise := packedEffectByte4B7540(motion, 24)
	center := image.Pt(int(int32(effect.Field_110)), int(int32(effect.Field_111)))

	direction := sincosTable16[angle]
	pointWorld := center.Add(image.Pt(radius*direction.X/16, radius*direction.Y/16))
	point := vp.ToScreenPos(pointWorld).Add(image.Pt(0, -int(int16(dr.ZVal))))
	state := vortexDrawState4B9F50{
		point:       point,
		inside:      point.X > vp.Screen.Min.X && point.X < vp.Screen.Max.X && point.Y > vp.Screen.Min.Y && point.Y < vp.Screen.Max.Y,
		nextAngle:   angle + byte(spin),
		nextRadius:  byte(radius),
		nextZ:       dr.ZVal + uint16(rise),
		pointColor:  effect.Field_108,
		glowSize:    3,
		pointRadius: 3,
	}
	if !state.inside {
		return state
	}
	if pointWorld.Y < center.Y {
		state.pointColor = vortexGray4B9F50
		state.glowSize = 2
		state.pointRadius = 2
	}

	tailAngle := byte(int(angle) - 2*int(spin))
	tailDirection := sincosTable16[tailAngle]
	tailWorld := center.Add(image.Pt(radius*tailDirection.X/16, radius*tailDirection.Y/16))
	state.tail = vp.ToScreenPos(tailWorld).Add(image.Pt(0, -int(int16(dr.ZVal))))
	state.tailColor = effect.Field_108
	if tailWorld.Y < center.Y {
		state.tailColor = vortexGray4B9F50
	}

	remaining := vortexRadiusForZ4B9F50(int16(state.nextZ))
	state.alive = remaining > 0
	if state.alive {
		state.nextRadius = byte(remaining)
	}
	return state
}

func applyVortexDrawState4B9F50(dr *client.Drawable, state vortexDrawState4B9F50) {
	effect := dr.UnionEffect()
	effect.Field_112 = setPackedEffectByte4B7540(effect.Field_112, 0, state.nextAngle)
	if state.alive {
		effect.Field_112 = setPackedEffectByte4B7540(effect.Field_112, 16, state.nextRadius)
	}
	dr.ZVal = state.nextZ
}

// VortexDraw previously indexed the Drawable through a truncated int. Keep
// its orbit, front/back colors, altitude decay, and strict viewport bounds
// while reading the native-width Drawable directly.
func (c *Client) drawVortex4B9F50(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := vortexDrawStateFor4B9F50(dr, vp)
	if !state.inside {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	pointColor := noxcolor.RGBA5551(state.pointColor)
	c.r.DrawGlow(state.point, pointColor, state.glowSize, state.glowSize+2)
	c.r.DrawPoint(state.point, state.pointRadius, pointColor)
	c.r.DrawLine(state.point, state.tail, noxcolor.RGBA5551(state.tailColor))
	applyVortexDrawState4B9F50(dr, state)
	if !state.alive {
		c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable(dr)
		return 0
	}
	return 1
}

//go:build !server

package opennox

import (
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

var (
	magicDrawWhite4B98A0        = noxcolor.RGB5551Color(255, 255, 255)
	magicDrawGlow4B98A0         = noxcolor.RGB5551Color(0, 200, 255)
	magicMissileGlow4B99F0      = noxcolor.RGB5551Color(255, 255, 50)
	magicDrawLight4B98A0        = noxrender.RGB{R: 200, G: 200, B: 255}
	magicMissileLight4B99F0     = noxrender.RGB{R: 255, G: 180, B: 50}
	magicTailLight4B5E10        = noxrender.RGB{R: 128, G: 128, B: 255}
	magicMissileTailLight4B5F30 = noxrender.RGB{R: 255, G: 128, B: 50}
)

type magicDrawableDrawKind4B98A0 uint8

const (
	magicDrawableDrawUnknown4B98A0 magicDrawableDrawKind4B98A0 = iota
	magicDrawableDrawOrb4B98A0
	magicDrawableDrawMissile4B99F0
	magicDrawableDrawTail4B5E10
	magicDrawableDrawMissileTail4B5F30
)

func magicDrawableDrawKindFor4B98A0(fn unsafe.Pointer) magicDrawableDrawKind4B98A0 {
	switch fn {
	case legacy.Get_nox_thing_magic_draw():
		return magicDrawableDrawOrb4B98A0
	case legacy.Get_nox_thing_magic_missle_draw():
		return magicDrawableDrawMissile4B99F0
	case legacy.Get_nox_thing_magic_tail_link_draw():
		return magicDrawableDrawTail4B5E10
	case legacy.Get_nox_thing_magic_missle_tail_link_draw():
		return magicDrawableDrawMissileTail4B5F30
	default:
		return magicDrawableDrawUnknown4B98A0
	}
}

func (c *Client) callMagicDrawableDraw4B98A0(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch magicDrawableDrawKindFor4B98A0(dr.DrawFuncPtr) {
	case magicDrawableDrawOrb4B98A0:
		return c.drawMagicOrb4B98A0(dr, vp, false), true
	case magicDrawableDrawMissile4B99F0:
		return c.drawMagicOrb4B98A0(dr, vp, true), true
	case magicDrawableDrawTail4B5E10:
		return c.drawMagicTail4B5E10(dr, vp, false), true
	case magicDrawableDrawMissileTail4B5F30:
		return c.drawMagicTail4B5E10(dr, vp, true), true
	default:
		return 0, false
	}
}

func magicDrawPoint4B98A0(dr *client.Drawable, vp *noxrender.Viewport) image.Point {
	return vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, -int(int16(dr.ZVal2))-int(int16(dr.ZVal))))
}

// MagicDraw and MagicMissileDraw used PE32 byte offsets after first copying
// the drawable pointer into an int. Draw and update their light natively.
func (c *Client) drawMagicOrb4B98A0(dr *client.Drawable, vp *noxrender.Viewport, missile bool) int {
	point := magicDrawPoint4B98A0(dr, vp)
	if point.X-10 < vp.Screen.Min.X || point.Y-10 < vp.Screen.Min.Y ||
		point.X+10 >= vp.Screen.Max.X || point.Y+10 >= vp.Screen.Max.Y {
		return 1
	}
	radius := c.srv.Rand.Other.Int(1, 4)
	glow, light := magicDrawGlow4B98A0, magicDrawLight4B98A0
	if missile {
		glow, light = magicMissileGlow4B99F0, magicMissileLight4B99F0
	}
	c.r.DrawGlow(point, glow, 2*radius+1, radius/2+3)
	c.r.Data().SetColor2(magicDrawWhite4B98A0)
	c.r.DrawRectFilledOpaque(point.X-radius/2, point.Y-radius/2, radius, radius, magicDrawWhite4B98A0)
	dr.SetLightColor(byte(light.R), byte(light.G), byte(light.B))
	dr.SetLightIntensity(float32(c.srv.Rand.Other.Float(0, 100)))
	return 1
}

type magicTailDrawState4B5E10 struct {
	point      image.Point
	tail       image.Point
	colorIndex int
	intensity  float32
	alive      bool
}

func magicTailDrawStateFor4B5E10(dr *client.Drawable, vp *noxrender.Viewport, frame, span uint32) magicTailDrawState4B5E10 {
	remaining := int32(dr.Deadline - frame)
	if remaining <= 0 || span == 0 {
		return magicTailDrawState4B5E10{}
	}
	effect := dr.UnionEffect()
	z := -int(int16(dr.ZVal2)) - int(int16(dr.ZVal))
	point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, z))
	tailWorld := image.Pt(int(int32(effect.Field_108)), int(int32(effect.Field_109)))
	tail := vp.ToScreenPos(tailWorld).Add(image.Pt(0, z))
	index := int(int64(remaining) * 64 / int64(span))
	if index >= 64 {
		index = 63
	}
	return magicTailDrawState4B5E10{
		point:      point,
		tail:       tail,
		colorIndex: index,
		intensity:  float32(float64(remaining) * 20 / float64(span)),
		alive:      true,
	}
}

// Magic tail callbacks read both the current position and effect-union tail
// endpoint through a truncated pointer. Their fade tables themselves remain
// valid fixed memory-map data, so only the drawable access is migrated.
func (c *Client) drawMagicTail4B5E10(dr *client.Drawable, vp *noxrender.Viewport, missile bool) int {
	span := c.srv.TickRate()
	colorBase := uintptr(1312500)
	light := magicTailLight4B5E10
	if missile {
		span /= 3
		colorBase = 1312756
		light = magicMissileTailLight4B5F30
	}
	state := magicTailDrawStateFor4B5E10(dr, vp, c.srv.Frame(), span)
	if !state.alive {
		return 1
	}
	color := noxcolor.RGBA5551(memmap.Uint32(0x5D4594, colorBase+uintptr(4*state.colorIndex)))
	dr.SetLightColor(byte(light.R), byte(light.G), byte(light.B))
	dr.SetLightIntensity(state.intensity)
	draw := c.r.Data()
	draw.SetColor2(color)
	draw.SetAlphaEnabled(true)
	draw.SetAlpha(0x80)
	c.r.DrawLine(state.point, state.tail, color)
	draw.SetAlphaEnabled(false)
	return 1
}

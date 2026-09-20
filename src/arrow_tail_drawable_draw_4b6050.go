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

type arrowTailDrawKind4B6050 uint8

const (
	arrowTailDrawUnknown4B6050 arrowTailDrawKind4B6050 = iota
	arrowTailDrawStrong4B6050
	arrowTailDrawWeak4B6120
)

func arrowTailDrawKindFor4B6050(fn unsafe.Pointer) arrowTailDrawKind4B6050 {
	switch fn {
	case legacy.Get_nox_thing_arrow_tail_link_draw():
		return arrowTailDrawStrong4B6050
	case legacy.Get_nox_thing_weak_arrow_tail_link_draw():
		return arrowTailDrawWeak4B6120
	default:
		return arrowTailDrawUnknown4B6050
	}
}

func (c *Client) callArrowTailDraw4B6050(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch arrowTailDrawKindFor4B6050(dr.DrawFuncPtr) {
	case arrowTailDrawStrong4B6050:
		return c.drawArrowTail4B6050(dr, vp, false), true
	case arrowTailDrawWeak4B6120:
		return c.drawArrowTail4B6050(dr, vp, true), true
	default:
		return 0, false
	}
}

type arrowTailDrawState4B6050 struct {
	point      image.Point
	tail       image.Point
	colorIndex int
	alive      bool
}

func arrowTailDrawStateFor4B6050(dr *client.Drawable, vp *noxrender.Viewport, frame, span uint32) arrowTailDrawState4B6050 {
	remaining := int32(dr.Deadline - frame)
	if remaining <= 0 || span == 0 {
		return arrowTailDrawState4B6050{}
	}
	effect := dr.UnionEffect()
	z := -int(int16(dr.ZVal2)) - int(int16(dr.ZVal)) - 4
	point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, z))
	tailWorld := image.Pt(int(int32(effect.Field_108)), int(int32(effect.Field_109)))
	tail := vp.ToScreenPos(tailWorld).Add(image.Pt(0, z))
	index := int(int64(remaining) * 64 / int64(span))
	if index >= 64 {
		index = 63
	}
	return arrowTailDrawState4B6050{
		point:      point,
		tail:       tail,
		colorIndex: index,
		alive:      true,
	}
}

// Arrow tail callbacks used PE32 byte offsets after copying the drawable
// pointer into an int. Read the current and previous positions through the
// native Drawable layout instead.
func (c *Client) drawArrowTail4B6050(dr *client.Drawable, vp *noxrender.Viewport, weak bool) int {
	state := arrowTailDrawStateFor4B6050(dr, vp, c.srv.Frame(), c.srv.TickRate()/3)
	if !state.alive {
		return 1
	}
	colorBase := uintptr(1313012)
	if weak {
		colorBase = 1313268
	}
	color := noxcolor.RGBA5551(memmap.Uint32(0x5D4594, colorBase+uintptr(4*state.colorIndex)))
	draw := c.r.Data()
	draw.SetColor2(color)
	draw.SetAlphaEnabled(true)
	draw.SetAlpha(0x80)
	c.r.DrawLine(state.point, state.tail, color)
	draw.SetAlphaEnabled(false)
	return 1
}

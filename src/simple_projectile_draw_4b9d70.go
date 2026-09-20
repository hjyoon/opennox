//go:build !server

package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

type simpleProjectileDrawKind4B9D70 uint8

const (
	simpleProjectileDrawUnknown4B9D70 simpleProjectileDrawKind4B9D70 = iota
	simpleProjectileDrawSpiderSpit4B9D70
	simpleProjectileDrawBlackPowder4B9ED0
)

func simpleProjectileDrawKindFor4B9D70(fn unsafe.Pointer) simpleProjectileDrawKind4B9D70 {
	switch fn {
	case legacy.Get_nox_thing_spider_spit_draw():
		return simpleProjectileDrawSpiderSpit4B9D70
	case legacy.Get_nox_thing_black_powder_draw():
		return simpleProjectileDrawBlackPowder4B9ED0
	default:
		return simpleProjectileDrawUnknown4B9D70
	}
}

func (c *Client) callSimpleProjectileDraw4B9D70(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	switch simpleProjectileDrawKindFor4B9D70(dr.DrawFuncPtr) {
	case simpleProjectileDrawSpiderSpit4B9D70:
		return c.drawSpiderSpit4B9D70(dr, vp), true
	case simpleProjectileDrawBlackPowder4B9ED0:
		return c.drawBlackPowder4B9ED0(dr, vp), true
	default:
		return 0, false
	}
}

func blackPowderPoint4B9ED0(dr *client.Drawable, vp *noxrender.Viewport) image.Point {
	return vp.ToScreenPos(dr.PosVec)
}

// BlackPowderDraw copied the drawable pointer into a 32-bit int before
// reading its position. Keep the pointer native-width and reproduce the four
// opaque powder marks directly through the Go renderer.
func (c *Client) drawBlackPowder4B9ED0(dr *client.Drawable, vp *noxrender.Viewport) int {
	point := blackPowderPoint4B9ED0(dr, vp)
	c.r.DrawRectFilledOpaque(point.X-1, point.Y-1, 3, 3, nox_color_gray2)
	c.r.DrawRectFilledOpaque(point.X-5, point.Y, 1, 1, nox_color_gray2)
	c.r.DrawRectFilledOpaque(point.X, point.Y+7, 1, 1, nox_color_gray2)
	c.r.DrawRectFilledOpaque(point.X+8, point.Y-2, 1, 1, nox_color_gray2)
	return 1
}

type projectileLine4B9D70 struct {
	from image.Point
	to   image.Point
}

type spiderSpitDrawState4B9D70 struct {
	point     image.Point
	tail      image.Point
	grayLines [2]projectileLine4B9D70
}

func spiderSpitDrawStateFor4B9D70(dr *client.Drawable, vp *noxrender.Viewport) spiderSpitDrawState4B9D70 {
	z := -int(int16(dr.ZVal))
	point := vp.ToScreenPos(dr.PosVec).Add(image.Pt(0, z))
	tailWorld := image.Pt(int(int32(dr.Field_8)), int(int32(dr.Field_9)))
	tail := vp.ToScreenPos(tailWorld).Add(image.Pt(0, z))

	dx := point.X - tail.X
	if dx < 0 {
		dx = -dx
	}
	dy := point.Y - tail.Y
	if dy < 0 {
		dy = -dy
	}
	state := spiderSpitDrawState4B9D70{point: point, tail: tail}
	if dx <= dy {
		state.grayLines[0] = projectileLine4B9D70{
			from: point.Add(image.Pt(1, 0)),
			to:   tail.Add(image.Pt(-1, 0)),
		}
		state.grayLines[1] = projectileLine4B9D70{
			from: point.Add(image.Pt(-1, 0)),
			to:   tail.Add(image.Pt(-1, 0)),
		}
	} else {
		state.grayLines[0] = projectileLine4B9D70{
			from: point.Add(image.Pt(0, 1)),
			to:   tail.Add(image.Pt(0, 1)),
		}
		state.grayLines[1] = projectileLine4B9D70{
			from: point.Add(image.Pt(0, 1)),
			to:   tail.Add(image.Pt(0, -1)),
		}
	}
	return state
}

// SpiderSpitDraw also truncated the drawable pointer before reading both the
// current and previous positions. Preserve its gray outline and white core.
func (c *Client) drawSpiderSpit4B9D70(dr *client.Drawable, vp *noxrender.Viewport) int {
	state := spiderSpitDrawStateFor4B9D70(dr, vp)
	for _, line := range state.grayLines {
		c.r.DrawLine(line.from, line.to, nox_color_gray2)
	}
	c.r.DrawRectFilledOpaque(state.point.X-1, state.point.Y-1, 4, 4, nox_color_gray2)
	c.r.DrawLine(state.point, state.tail, nox_color_white_2523948)
	c.r.DrawRectFilledOpaque(state.point.X, state.point.Y, 2, 2, nox_color_white_2523948)
	return 1
}

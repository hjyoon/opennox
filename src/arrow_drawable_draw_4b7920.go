//go:build !server

package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

type arrowDrawKind4B7920 uint8

const (
	arrowDrawUnknown4B7920 arrowDrawKind4B7920 = iota
	arrowDrawStrong4B7920
	arrowDrawWeak4B79D0
)

func arrowDrawKindFor4B7920(fn unsafe.Pointer) arrowDrawKind4B7920 {
	switch fn {
	case legacy.Get_nox_thing_arrow_draw():
		return arrowDrawStrong4B7920
	case legacy.Get_nox_thing_weak_arrow_draw():
		return arrowDrawWeak4B79D0
	default:
		return arrowDrawUnknown4B7920
	}
}

type arrowDrawHooks4B7920 struct {
	typeID    func(string) int
	spawn     func(int, image.Point) *client.Drawable
	activate  func(*client.Drawable)
	decay     func(*client.Drawable, int)
	slaveDraw func(*noxrender.Viewport, *client.Drawable) int
	decayTime int
}

// drawArrow4B7920 replaces the PE32 callback, which treated a native-width
// Drawable as []uint32. On 64-bit hosts its write to index 82 lands on the
// low half of DrawFuncPtr instead of Field_82 and corrupts the next draw call.
func drawArrow4B7920(dr *client.Drawable, vp *noxrender.Viewport, tailType string, hooks arrowDrawHooks4B7920) int {
	previous := image.Pt(int(int32(dr.Field_81)), int(int32(dr.Field_82)))
	delta := dr.PosVec.Sub(previous)
	distanceSquared := int64(delta.X)*int64(delta.X) + int64(delta.Y)*int64(delta.Y)
	if distanceSquared > 200 {
		tail := hooks.spawn(hooks.typeID(tailType), previous)
		if tail != nil {
			effect := tail.UnionEffect()
			effect.Field_108 = uint32(dr.PosVec.X)
			effect.Field_109 = uint32(dr.PosVec.Y)
			hooks.activate(tail)
			dr.Field_81 = uint32(dr.PosVec.X)
			dr.Field_82 = uint32(dr.PosVec.Y)
			hooks.decay(tail, hooks.decayTime)
		}
	}
	return hooks.slaveDraw(vp, dr)
}

func (c *Client) callArrowDraw4B7920(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	var tailType string
	switch arrowDrawKindFor4B7920(dr.DrawFuncPtr) {
	case arrowDrawStrong4B7920:
		tailType = "ArrowTailLink"
	case arrowDrawWeak4B79D0:
		tailType = "WeakArrowTailLink"
	default:
		return 0, false
	}
	return drawArrow4B7920(dr, vp, tailType, arrowDrawHooks4B7920{
		typeID:    c.Things.IndByID,
		spawn:     c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		activate:  c.Objs.List34Add,
		decay:     c.Objs.TransparentDecay,
		slaveDraw: legacy.Nox_thing_slave_draw,
		decayTime: int(c.srv.TickRate() / 3),
	}), true
}

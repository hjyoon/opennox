//go:build !server

package opennox

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
)

// summonAnimateDrawData4B7D00 mirrors nox_animate_draw_data_t. Go's native
// pointer alignment intentionally matches the C parser on both 32- and
// 64-bit targets.
type summonAnimateDrawData4B7D00 struct {
	size          uint32
	images        unsafe.Pointer
	frameCount    uint8
	frameDelay    uint8
	animationKind uint32
}

type summonEffectDrawHooks4B7D00 struct {
	frame       func() uint32
	animate     func(*client.Drawable) int
	drawChild   func(*client.Drawable, byte) int
	pointSpark  func(image.Point)
	trig        func(int) image.Point
	deleteChild func(*client.Drawable)
	deleteSelf  func(*client.Drawable)
}

func drawSummonEffect4B7D00(dr *client.Drawable, hooks summonEffectDrawHooks4B7D00) int {
	if dr == nil {
		return 0
	}
	now := hooks.frame()
	data := dr.UnionSummon()
	elapsed := now - dr.AnimStart
	lifetime := uint32(data.Lifetime)
	if elapsed >= lifetime {
		if data.Child != nil {
			hooks.deleteChild(data.Child)
			data.Child = nil
		}
		hooks.deleteSelf(dr)
		return 0
	}
	if elapsed >= lifetime-1 {
		hooks.pointSpark(dr.PosVec)
	}

	_ = hooks.animate(dr)
	anim := (*summonAnimateDrawData4B7D00)(dr.DrawData)
	if anim != nil && anim.frameCount != 0 {
		originalPos := dr.PosVec
		originalFrame := dr.AnimFrameSlave
		phase := uint32(0)
		for i := 0; i < 26; i++ {
			if phase >= uint32(anim.frameCount) {
				phase = 0
			}
			frame := (phase + now + dr.NetCode32) / (uint32(anim.frameDelay) + 1)
			dr.AnimFrameSlave = frame % uint32(anim.frameCount)
			anim.animationKind = 5
			off := hooks.trig(i)
			dr.PosVec = originalPos.Add(image.Pt(2*off.X, 2*off.Y))
			if dr.PosVec.X >= 0 && dr.PosVec.X < 5888 && dr.PosVec.Y >= 0 && dr.PosVec.Y < 5888 {
				_ = hooks.animate(dr)
			}
			phase++
		}
		dr.PosVec = originalPos
		dr.AnimFrameSlave = originalFrame
		anim.animationKind = 2
	}

	if data.Child != nil {
		alpha := byte(elapsed * 255 / lifetime)
		_ = hooks.drawChild(data.Child, alpha)
	}
	return 1
}

func (c *Client) drawSummonEffect4B7D00(dr *client.Drawable, vp *noxrender.Viewport) int {
	return drawSummonEffect4B7D00(dr, summonEffectDrawHooks4B7D00{
		frame: c.srv.Frame,
		animate: func(got *client.Drawable) int {
			return legacy.Nox_thing_animate_draw(vp, got)
		},
		drawChild: func(child *client.Drawable, alpha byte) int {
			draw := c.r.Data()
			draw.SetAlphaEnabled(true)
			draw.SetAlpha(alpha)
			res := c.callDrawableDraw4B6B80(child, vp)
			draw.SetAlphaEnabled(false)
			return res
		},
		pointSpark: func(pos image.Point) {
			c.spawnPointSparkFXNative48EA70(pointSparkFXSpec48EA70{cache: 0, name: "BlueSpark", count: 50, speed: 1000, minLife: 30}, pos)
		},
		trig: func(i int) image.Point {
			off := uintptr(192092 + 80*i)
			return image.Pt(
				int(memmap.Int32(0x587000, off-4)),
				int(memmap.Int32(0x587000, off)),
			)
		},
		deleteChild: func(child *client.Drawable) {
			c.Nox_xxx_spriteDelete_45A4B0(child)
		},
		deleteSelf: c.Nox_xxx_spriteDeleteStatic_45A4E0_drawable,
	})
}

func (c *Client) callSummonEffectDraw4B7D00(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_summon_effect_draw() {
		return 0, false
	}
	return c.drawSummonEffect4B7D00(dr, vp), true
}

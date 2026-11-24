package legacy

/*
#include "GAME2_3.h"
*/
import "C"

import (
	"image"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func drawableFXHead() **client.DrawableFX {
	return (**client.DrawableFX)(unsafe.Pointer(memmap.PtrPtr(0x5D4594, 1203872)))
}

func linkDrawableFX(fx *client.DrawableFX, dr *client.Drawable) {
	if fx == nil || dr == nil {
		return
	}
	head := drawableFXHead()
	fx.Owner = dr
	fx.GlobalNext, fx.GlobalPrev = *head, nil
	if fx.GlobalNext != nil {
		fx.GlobalNext.GlobalPrev = fx
	}
	*head = fx
	fx.Next, fx.Prev = dr.Field_114, nil
	if fx.Next != nil {
		fx.Next.Prev = fx
	}
	dr.Field_114 = fx
}

func unlinkDrawableFX(fx *client.DrawableFX) {
	if fx == nil {
		return
	}
	if fx.GlobalNext != nil {
		fx.GlobalNext.GlobalPrev = fx.GlobalPrev
	}
	if fx.GlobalPrev != nil {
		fx.GlobalPrev.GlobalNext = fx.GlobalNext
	} else {
		*drawableFXHead() = fx.GlobalNext
	}
	if fx.Next != nil {
		fx.Next.Prev = fx.Prev
	}
	if fx.Prev != nil {
		fx.Prev.Next = fx.Next
	} else if fx.Owner != nil {
		fx.Owner.Field_114 = fx.Next
	}
	fx.Owner, fx.Next, fx.Prev, fx.GlobalNext, fx.GlobalPrev = nil, nil, nil, nil, nil
}

func drawDrawableTrails(dr *client.Drawable, vp *noxrender.Viewport) {
	if dr == nil || vp == nil {
		return
	}
	for fx := dr.Field_114; fx != nil; {
		next := fx.Next
		switch fx.Field0 {
		case 1:
			drawSpriteTrail(dr, fx, vp)
		case 2:
			drawLineTrail(dr, fx, vp)
		}
		fx = next
	}
}

func drawableTrailLength(dr *client.Drawable, fx *client.DrawableFX) (int, bool) {
	count := min(int(fx.Count), len(fx.Trail)-1)
	for i := range count {
		// Axis-aligned pairs count as collapsed in GAME.EXE.
		if fx.Trail[i][0] != fx.Trail[i+1][0] && fx.Trail[i][1] != fx.Trail[i+1][1] {
			return count, true
		}
	}
	if int32(dr.PosVec.X) != int32(dr.Field_8) || int32(dr.PosVec.Y) != int32(dr.Field_9) {
		return count, true
	}
	fx.Count = 0
	return count, false
}

func advanceDrawableTrail(fx *client.DrawableFX, point image.Point, count int) {
	copy(fx.Trail[1:count+1], fx.Trail[:count])
	fx.Trail[0] = [2]int32{int32(point.X), int32(point.Y)}
	fx.Count = uint8(min(count+1, len(fx.Trail)-1))
}

func drawSpriteTrail(dr *client.Drawable, fx *client.DrawableFX, vp *noxrender.Viewport) int {
	if dr == nil || fx == nil || vp == nil {
		return 0
	}
	count, draw := drawableTrailLength(dr, fx)
	if !draw {
		return count
	}
	position := dr.Pos()
	data := GetClient().R2().Data()
	alpha := uint8(255)
	for i := range count {
		alpha -= 42
		data.SetAlphaEnabled(true)
		data.SetAlpha(alpha)
		dr.SetPos(image.Pt(int(fx.Trail[i][0]), int(fx.Trail[i][1])))
		dr.CallDraw(vp)
	}
	dr.SetPos(position)
	advanceDrawableTrail(fx, position, count)
	data.SetAlphaEnabled(false)
	return 1
}

func drawLineTrail(dr *client.Drawable, fx *client.DrawableFX, vp *noxrender.Viewport) int {
	if dr == nil || fx == nil || vp == nil {
		return 0
	}
	count, draw := drawableTrailLength(dr, fx)
	if !draw {
		return count
	}
	velocity := memmap.PtrT[[2]float32](0x587000, uintptr(194136+64*dr.AnimFrameSlave))
	dx := int32(vp.Screen.Min.X-vp.World.Min.X) + server.X87TruncateSignedDWORD419A70(-12*velocity[0])
	dy := int32(vp.Screen.Min.Y-vp.World.Min.Y) - int32(int16(dr.ZVal)) - int32(int16(dr.ZVal2)) - 10 +
		server.X87TruncateSignedDWORD419A70(-12*velocity[1])
	previous := image.Pt(int(int32(dr.PosVec.X)+dx), int(int32(dr.PosVec.Y)+dy))
	r := GetClient().R2()
	data := r.Data()
	color := noxcolor.RGB5551Color(200, 200, 200)
	alpha := uint8(255)
	for i := range count {
		alpha -= 63
		data.SetAlphaEnabled(true)
		data.SetAlpha(alpha)
		current := image.Pt(int(fx.Trail[i][0]+dx), int(fx.Trail[i][1]+dy))
		data.SetColor2(color)
		r.AddPoint(previous)
		r.AddPoint(current)
		r.DrawLineFromPoints(data.Color2())
		previous = current
	}
	advanceDrawableTrail(fx, dr.Pos(), count)
	data.SetAlphaEnabled(false)
	return 1
}

//export sub_495B50
func sub_495B50(fx *C.nox_drawable_fx) {
	unlinkDrawableFX((*client.DrawableFX)(unsafe.Pointer(fx)))
}

//export sub_495BB0
func sub_495BB0(dr *nox_drawable, view *nox_draw_viewport_t) {
	drawDrawableTrails(asDrawable(dr), asViewport(view))
}

//export sub_495BF0
func sub_495BF0(dr *nox_drawable, fx *C.nox_drawable_fx, view *nox_draw_viewport_t) C.int {
	return C.int(drawSpriteTrail(asDrawable(dr), (*client.DrawableFX)(unsafe.Pointer(fx)), asViewport(view)))
}

//export sub_495D00
func sub_495D00(dr *nox_drawable, fx *C.nox_drawable_fx, view *nox_draw_viewport_t) C.int {
	return C.int(drawLineTrail(asDrawable(dr), (*client.DrawableFX)(unsafe.Pointer(fx)), asViewport(view)))
}

//export sub_495FC0
func sub_495FC0(fx *C.nox_drawable_fx, dr *nox_drawable) {
	linkDrawableFX((*client.DrawableFX)(unsafe.Pointer(fx)), asDrawable(dr))
}

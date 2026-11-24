package legacy

/*
#include "GAME1_2.h"
#include "GAME3_1.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_client_drawable_testBuff_4356C0
func nox_client_drawable_testBuff_4356C0(dr *nox_drawable, id C.char) C.bool {
	return C.bool(asDrawable(dr).HasEnchant(server.EnchantID(id)))
}

//export sub_4C4EC0
func sub_4C4EC0(vp *nox_draw_viewport_t, dr *nox_drawable) C.char {
	observer := AsDrawableP(*memmap.PtrPtr(0x852978, 8))
	if observer.HasEnchant(server.ENCHANT_INFRAVISION) {
		return ^C.char(0)
	}
	view := (*noxrender.Viewport)(unsafe.Pointer(vp))
	pos := asDrawable(dr).Pos()
	x := int32(pos.X) - int32(view.Size.X)/2 - int32(view.World.Min.X)
	y := int32(pos.Y) - int32(view.Size.Y)/2 - int32(view.World.Min.Y)
	divisor := int32(memmap.Uint32(0x587000, 185464))
	if divisor == 0 {
		return 0
	}
	alpha := 128 - ((x*x+y*y)<<7)/divisor
	return C.char(uint8(min(128, max(0, alpha))))
}

//export nox_xxx_drawShinySpot_4C4F40
func nox_xxx_drawShinySpot_4C4F40(vp *nox_draw_viewport_t, dr *nox_drawable) C.short {
	return drawShinySpot4C4F40((*noxrender.Viewport)(unsafe.Pointer(vp)), asDrawable(dr))
}

func drawShinySpot4C4F40(vp *noxrender.Viewport, dr *client.Drawable) C.short {
	slot := memmap.PtrPtr(0x5D4594, 1321524)
	ref := AsImageRefP(*slot)
	if ref == nil {
		ref = Nox_xxx_gLoadAnim("ShinySpot")
		*slot = ref.C()
	}
	anim := ref.Field24ptr()
	count := uint32(anim.ImagesSz)
	period := 8 * count
	elapsed := gameFrame() + dr.Field32()
	frame := (elapsed % period) >> 1
	if frame < count {
		screen := vp.ToScreenPos(dr.Pos())
		pos := image.Pt(int(int32(screen.X)-64), int(int32(screen.Y)-int32(dr.ZVal2)-int32(dr.Z())-64))
		render := GetClient().R2()
		render.DrawImageAt(render.GetBag().AsImage(anim.Images()[frame]), pos)
	}
	return C.short(elapsed / period)
}

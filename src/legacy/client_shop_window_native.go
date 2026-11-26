package legacy

/*
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "client__gui__guishop.h"
#include "client__gui__window.h"

typedef const nox_draw_viewport_t nox_shop_const_viewport_t;

*/
import "C"

import (
	"image"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

var shopWindow, shopSlider *nox_window
var shopBorderImage *nox_video_bag_image_t
var shopScrollLimit C.uint32_t

var shopDrawViewport, _ = alloc.New(C.nox_draw_viewport_t{})

var shopImages struct {
	base, trade, identify, repair *noxrender.Image
	rows                          [2]*noxrender.Image
}

//export nox_client_shop_viewport_init
func nox_client_shop_viewport_init(width, height C.int) {
	w, h := C.intptr_t(width), C.intptr_t(height)
	*shopDrawViewport = C.nox_draw_viewport_t{x2: w, y2: h, width: w, height: h}
}

//export nox_client_shop_viewport
func nox_client_shop_viewport() *C.nox_shop_const_viewport_t {
	return (*C.nox_shop_const_viewport_t)(shopDrawViewport)
}

//export sub_478110
func sub_478110() C.int {
	nox_client_shop_viewport_init(clientDisplayStorage.width, clientDisplayStorage.height)
	root := nox_new_window_from_file(internCStr("Shop.wnd"), C.sub_478480)
	shopWindow, shopSlider = root, nil
	if root == nil {
		return 0
	}
	nox_window_set_all_funcs(root, C.sub_478650, C.sub_478970, nil)
	inventory := nox_xxx_wndGetChildByID_46B0C0(root, 3806)
	slider := nox_xxx_wndGetChildByID_46B0C0(root, 3807)
	if inventory == nil || slider == nil || slider.field_100 == nil || slider.widget_data == nil {
		nox_xxx_windowDestroyMB_46C4E0(root)
		shopWindow = nil
		return 0
	}
	asWindow(root).AddCleanup(func() {
		if shopWindow == root {
			shopWindow, shopSlider = nil, nil
		}
	})
	nox_gui_winSetFunc96_46B070(inventory, C.sub_478E50)
	// The game clears startup storage before paths that can destroy drawables.
	for i := range shopInventory {
		shopInventory[i].drawable = nil
	}
	sub_478F10()
	win := asWindow(root)
	win.SetPos(image.Pt(int(clientDisplayStorage.width), int(clientDisplayStorage.height)).Sub(win.Size()))
	win.Hide()
	win.SetEnabled(false)

	shopImages.base = Nox_xxx_gLoadImg("ShopBase")
	shopImages.trade = Nox_xxx_gLoadImg("ShopTradeMode")
	shopImages.identify = Nox_xxx_gLoadImg("ShopIdentifyMode")
	shopImages.repair = Nox_xxx_gLoadImg("ShopRepairMode")
	Nox_xxx_gLoadImg("ShopExitMode")
	shopImages.rows = [2]*noxrender.Image{Nox_xxx_gLoadImg("ShopInventoryBar1"), Nox_xxx_gLoadImg("ShopInventoryBar2")}
	thumb := Nox_xxx_gLoadImg("ShopInventorySlider")
	thumbPressed := Nox_xxx_gLoadImg("ShopInventorySliderSelected")
	for _, name := range [...]string{"ShopInventoryUp", "ShopInventoryUpSelected", "ShopInventorydown", "ShopInventorydownSelected"} {
		Nox_xxx_gLoadImg(name)
	}
	shopBorderImage = (*nox_video_bag_image_t)(Nox_xxx_gLoadImg("ShopTextBorder").C())
	for _, entry := range shopPortraits[:6] {
		Nox_xxx_gLoadImg(entry.image)
	}
	shopSlider = slider
	origin := asWindow(inventory).GlobalPos()
	end := origin.Add(asWindow(inventory).Size())
	*g2ServerWord(1098380), *g2ServerWord(1098384) = uint32(origin.X), uint32(origin.Y)
	*g2ServerWord(1098388), *g2ServerWord(1098392) = uint32(end.X), uint32(end.Y)
	slider.field_100.width, slider.field_100.height = 16, 12
	nox_xxx_wndSetOffsetMB_46AE40(slider.field_100, 0, -15)
	C.sub_4B5700(slider, nil, nil, unsafe.Pointer(thumb.C()), unsafe.Pointer(thumbPressed.C()), unsafe.Pointer(thumbPressed.C()))
	shopScrollLimit = (*[2]C.uint32_t)(slider.widget_data)[1]
	return 1
}

//export sub_478480
func sub_478480(win *nox_window, event C.int, arg, value C.uintptr_t) C.int {
	if event == 16393 {
		clientItemStorage.shop_scroll = shopScrollLimit - C.uint32_t(value)
	}
	if event != 16391 || sub_45D9B0() != 0 {
		return 0
	}
	id := nox_xxx_wndGetID_46B0A0((*nox_window)(unsafe.Pointer(uintptr(arg))))
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 3801:
		sub_478040()
	case 3802:
		if clientUIWords.shop_active != 0 {
			if clientUIWords.shop_mode == 4 {
				sub_467680()
			}
			nox_client_setCursorType_477610(12)
			clientUIWords.shop_mode = 3
			if C.sub_467C80() == 0 {
				C.sub_467BB0()
			}
		}
	case 3803:
		if clientUIWords.shop_active != 0 {
			C.sub_467650()
			clientUIWords.shop_mode = 4
		}
	case 3804:
		if clientUIWords.shop_active != 0 {
			if clientUIWords.shop_mode == 4 {
				sub_467680()
			}
			nox_client_setCursorType_477610(11)
			clientUIWords.shop_mode = 2
		}
	case 3808:
		offset := clientItemStorage.shop_scroll - 50
		if int32(offset) < 0 {
			offset = 0
		}
		shopSetScrollOffset(offset - offset%50)
	case 3809:
		offset := clientItemStorage.shop_scroll + 50
		if int32(offset) <= int32(shopScrollLimit) {
			shopSetScrollOffset(offset - offset%50)
		} else {
			shopSetScrollOffset(shopScrollLimit)
		}
	}
	return 0
}

func shopSetScrollOffset(offset C.uint32_t) {
	clientItemStorage.shop_scroll = offset
	g2ServerEvent(shopSlider, 16394, uintptr(shopScrollLimit-offset), 0)
}

func shopMousePosition(packed C.uintptr_t) C.int2 {
	return C.int2{field_0: C.int(uint16(packed)), field_4: C.int(uint32(packed) >> 16)}
}

//export sub_478650
func sub_478650(win *nox_window, event C.int, arg, value C.uintptr_t) C.int {
	if sub_45D9B0() != 0 {
		return 1
	}
	switch event {
	case 5:
		pos := shopMousePosition(arg)
		if clientUIWords.shop_mode == 2 && sub_479880((*C.uint32_t)(unsafe.Pointer(&pos))) {
			sub_478730(&pos.field_0)
		}
	case 19, 20:
		if clientUIWords.shop_mode == 2 {
			button := nox_xxx_wndGetChildByID_46B0C0(shopRoot(), 3808+int32(event)-19)
			g2ServerText(shopRoot(), 16391, unsafe.Pointer(button), 0)
		}
	default:
		return 0
	}
	return 1
}

//export sub_478970
func sub_478970(win *nox_window, draw *C.nox_window_data) C.int {
	x := clientDisplayStorage.width - C.NOX_DEFAULT_WIDTH
	y := clientDisplayStorage.height - C.NOX_DEFAULT_HEIGHT
	pos := C.int2{field_0: x, field_4: y}
	r := GetClient().R2()
	at := image.Pt(int(x), int(y))
	r.DrawImageAt(shopImages.base, at)
	switch clientUIWords.shop_mode {
	case 2:
		r.DrawImageAt(shopImages.trade, at)
		sub_478C80()
	case 3:
		r.DrawImageAt(shopImages.identify, at)
		sub_478B10(&pos)
	case 1:
		r.DrawImageAt(shopImages.repair, at)
		sub_478A70(&pos)
	case 4:
		r.DrawImageAt(shopImages.repair, at)
		sub_478BC0(&pos.field_0)
	}
	return 1
}

//export sub_478C80
func sub_478C80() C.int {
	gold := uint32(sub_4674A0())
	nox_xxx_wndDraw_49F7F0()
	x0, y0 := int(int32(*g2ServerWord(1098380))), int(int32(*g2ServerWord(1098384)))
	x1, y1 := int(int32(*g2ServerWord(1098388))), int(int32(*g2ServerWord(1098392)))
	currentDrawClipRect(C.int(x0), C.int(y0), C.int(x1-x0), C.int(y1-y0))
	r := GetClient().R2()
	fontHeight := r.FontHeight(nil)
	view := asViewport(shopDrawViewport)
	y := y0 - int(int32(clientItemStorage.shop_scroll))
	for row := 0; row < shopRows && y < y1; row, y = row+1, y+50 {
		if y <= y0-50 {
			continue
		}
		r.DrawImageAt(shopImages.rows[row%2], image.Pt(x0, y))
		for col := range shopCols {
			cell := shopCell(row, col)
			if cell.count == 0 {
				continue
			}
			x := x0 + 5 + 50*col
			drawable := asDrawable(cell.drawable)
			drawable.SetPos(image.Pt(x+20, y+25))
			drawable.CallDraw(view)
			if gold < uint32(cell.price) {
				nox_client_drawRectFilledAlpha_49CF10(int32(x-5), int32(y), 50, 50)
			}
			if cell.count > 1 {
				nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
				r.DrawStringWrapped(nil, strconv.FormatInt(int64(int32(cell.count)), 10), image.Rect(x, y+5, x+320, y+5))
			}
			lineY := y - fontHeight + 45
			nox_xxx_drawSetTextColor_434390(int32(renderPalette.yellow))
			r.DrawStringWrapped(nil, strconv.FormatInt(int64(int32(cell.price)), 10), image.Rect(x, lineY, x+320, lineY))
		}
	}
	return sub_49F860()
}

//export sub_478E50
func sub_478E50(win *nox_window, draw *C.nox_window_data, arg C.uintptr_t) C.int {
	if clientUIWords.shop_mode == 2 {
		pos := shopMousePosition(arg)
		col := (uint32(pos.field_0) - *g2ServerWord(1098380)) / 50
		row := int32(uint32(pos.field_4)-*g2ServerWord(1098384)+uint32(clientItemStorage.shop_scroll)) / 50
		cell := shopCell(int(max(0, min(row, shopRows-1))), int(min(col, shopCols-1)))
		if cell.count != 0 {
			cell.drawable.field_32 = cell.net_codes[0]
			nox_xxx_cursorSetTooltip_4776B0(nox_xxx_clientAskInfoMb_4BF050(cell.drawable))
		}
	}
	return 1
}

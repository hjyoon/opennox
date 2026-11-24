package legacy

/*
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__gui__window.h"
#include "noxstring.h"
extern int nox_win_width, nox_win_height;
extern nox_window* dword_5d4594_1320940;
extern uint32_t dword_5d4594_1320944, dword_5d4594_1320948;
extern uint32_t nox_color_white_2523948, nox_color_yellow_2589772;
*/
import "C"

import (
	"image"
	"strconv"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const tradeCellSize = 50

func tradeRoot() *nox_window { return C.dword_5d4594_1320940 }

func tradeChild(id int) *nox_window { return g2ServerChild(tradeRoot(), id) }

func tradeImage(offset uintptr) *nox_video_bag_image_t {
	return (*nox_video_bag_image_t)(*memmap.PtrPtr(0x5D4594, offset))
}

//export sub_4C09D0
func sub_4C09D0() C.int {
	root := nox_new_window_from_file(internCStr("Trade.wnd"), C.sub_4C0C90)
	C.dword_5d4594_1320940 = root
	if root == nil {
		return 0
	}
	nox_window_set_all_funcs_go(root, C.sub_4C0630, C.sub_4C0D00, nil)
	nox_xxx_wndWddSetTooltip_46B000(&root.draw_data, tradeString("TradeMain", 692))
	for _, tip := range [...]struct {
		id   int
		name string
		line C.int
	}{
		{3702, "TradePlayerName", 695},
		{3703, "TradeVendorName", 698},
	} {
		nox_xxx_wndWddSetTooltip_46B000(&tradeChild(tip.id).draw_data, tradeString(tip.name, tip.line))
	}
	nox_gui_winSetFunc96_46B070(tradeChild(3704), C.sub_4C1120)
	nox_gui_winSetFunc96_46B070(tradeChild(3705), C.sub_4C1120)
	for _, tip := range [...]struct {
		id   int
		name string
		line C.int
	}{
		{3708, "TradePlayerAccept", 709},
		{3709, "TradeVendorAccept", 712},
		{3710, "TradeCancel", 715},
	} {
		nox_xxx_wndWddSetTooltip_46B000(&tradeChild(tip.id).draw_data, tradeString(tip.name, tip.line))
	}
	g2ServerHidden(root, true)
	g2ServerEnabled(root, false)
	nox_gui_trade_slots_reset()
	C.nox_wcscpy((*wchar2_t)(g2ServerMem(1319972)), tradeString("TotalValueLabel", 749))
	*g2ServerWord(1320188), *g2ServerWord(1320192) = 0, 0
	*g2ServerWord(1320196), *g2ServerWord(1320200) = uint32(C.nox_win_width), uint32(C.nox_win_height)
	*g2ServerWord(1320220), *g2ServerWord(1320224) = uint32(C.nox_win_width), uint32(C.nox_win_height)
	*g2ServerWord(1320204), *g2ServerWord(1320208) = 0, 0
	for i, name := range [...]string{
		"TradeBase", "TradeLeftAcceptPushed", "TradeLeftAcceptLit", "TradeRightAcceptLit", "TradeCancelLit", "TradeGold",
	} {
		*memmap.PtrPtr(0x5D4594, 1320164+4*uintptr(i)) = unsafe.Pointer(Nox_xxx_gLoadImg(name).C())
	}
	return 1
}

func drawTradeSlots(remote C.int, origin image.Point, fontHeight int32) {
	for row := C.int(0); row < 2; row++ {
		for col := C.int(0); col < 2; col++ {
			slot := nox_gui_trade_slot_at(remote, row+2*col)
			if slot.count == 0 || slot.drawable == nil {
				continue
			}
			dr := asDrawable(slot.drawable)
			x, y := int32(origin.X)+int32(col)*tradeCellSize, int32(origin.Y)+int32(row)*tradeCellSize
			dr.SetPos(image.Pt(int(x+25), int(y+25)))
			viewport := packedGUIViewport(1320188)
			dr.CallDraw(&viewport)
			var buffer [32]uint16
			text := (*wchar2_t)(unsafe.Pointer(&buffer[0]))
			WStrCopy(text, len(buffer), strconv.FormatInt(int64(int32(slot.count)), 10))
			nox_xxx_drawSetTextColor_434390(int32(C.nox_color_white_2523948))
			nox_xxx_drawString_43F6E0(nil, text, x+5, y+5)
			WStrCopy(text, len(buffer), strconv.FormatInt(int64(int32(slot.total_cost)), 10))
			nox_xxx_drawSetTextColor_434390(int32(C.nox_color_yellow_2589772))
			nox_xxx_drawString_43F6E0(nil, text, x+5, y+45-fontHeight)
		}
	}
}

//export sub_4C0D00
func sub_4C0D00(win *nox_window, draw *C.nox_window_data) C.int {
	pos := asWindow(tradeRoot()).Offs()
	x, y := int32(pos.X), int32(pos.Y)
	drawImage := func(offset uintptr) {
		nox_client_drawImageAt_47D2C0(tradeImage(offset), x, y)
	}
	drawImage(1320164)
	g2ServerText(tradeChild(3711), 16385, g2ServerMem(1320240), 0)
	g2ServerText(tradeChild(3712), 16385, g2ServerMem(1320868), 0)
	g2ServerText(tradeChild(3713), 16385, g2ServerMem(1320100), 0)
	for _, offset := range [...]uintptr{183696, 183704} {
		gx := x + memmap.Int32(0x587000, offset) - 64
		gy := y + memmap.Int32(0x587000, offset+4) - 64
		nox_client_drawImageAt_47D2C0(tradeImage(1320184), gx, gy)
	}
	accept := tradeChild(3708)
	if C.dword_5d4594_1320944 != 0 {
		drawImage(1320172)
	} else if accept.draw_data.field_0&4 != 0 {
		drawImage(1320168)
	}
	if C.dword_5d4594_1320948 != 0 {
		drawImage(1320176)
	}
	cancel := tradeChild(3710)
	if *g2ServerWord(1320960) != 0 || cancel.draw_data.field_0&4 != 0 {
		drawImage(1320180)
	}
	local := asWindow(tradeChild(3704)).GlobalPos()
	height := nox_xxx_guiFontHeightMB_43F320(nil)
	drawTradeSlots(0, local, height)
	remote := asWindow(tradeChild(3705)).GlobalPos()
	drawTradeSlots(1, remote, height)
	return 1
}

func tradeSlotFromPosition(remote C.int, pos *C.int2) *tradeClientSlot {
	origin := asWindow(tradeChild(3704 + int(remote))).GlobalPos()
	for row := C.int(0); row < 2; row++ {
		for col := C.int(0); col < 2; col++ {
			x, y := C.int(origin.X)+tradeCellSize*col, C.int(origin.Y)+tradeCellSize*row
			bounds := C.int4{field_0: x, field_4: y, field_8: x + tradeCellSize, field_C: y + tradeCellSize}
			if nox_xxx_pointInRect_4281F0(pos, &bounds) != 0 {
				return nox_gui_trade_slot_at(remote, row+2*col)
			}
		}
	}
	return nil
}

//export sub_4C0910
func sub_4C0910(pos *C.int2) *tradeClientSlot { return tradeSlotFromPosition(0, pos) }

//export sub_4C11E0
func sub_4C11E0(pos *C.int2) *tradeClientSlot { return tradeSlotFromPosition(1, pos) }

//export sub_4C1120
func sub_4C1120(win *nox_window, draw *C.nox_window_data, packed C.int) C.int {
	pos := C.int2{field_0: C.int(uint16(packed)), field_4: C.int(uint32(packed) >> 16)}
	var slot *tradeClientSlot
	if C.nox_xxx_wndPointInWnd_46AAB0(tradeChild(3704), pos.field_0, pos.field_4) {
		slot = sub_4C0910(&pos)
	} else if C.nox_xxx_wndPointInWnd_46AAB0(tradeChild(3705), pos.field_0, pos.field_4) {
		slot = sub_4C11E0(&pos)
	}
	if slot != nil && slot.count != 0 && slot.drawable != nil {
		dr := slot.drawable
		dr.field_32 = slot.item_ids[0]
		C.nox_xxx_cursorSetTooltip_4776B0(nox_xxx_clientAskInfoMb_4BF050(dr))
	}
	return 1
}

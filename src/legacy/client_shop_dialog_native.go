package legacy

/*
#include "GAME1_3.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__gui__guishop.h"
#include "noxstring.h"

typedef const wchar2_t nox_shop_const_wchar2_t;

extern nox_window* dword_5d4594_1098580;
extern uint32_t dword_5d4594_1098592, dword_5d4594_1098616, dword_5d4594_1098620;
extern uint32_t dword_5d4594_1098624, dword_5d4594_1098628, dword_5d4594_1107036;
extern wchar2_t *dword_5d4594_1098596, *dword_5d4594_1098600, *dword_5d4594_1098604;
extern char* nox_shop_dialog_file_1098608;
extern nox_video_bag_image_t* dword_5d4594_1098456;
extern uint32_t nox_color_white_2523948;

static inline void shop_format_gold_shortage(wchar2_t* dst, const wchar2_t* format, int amount) {
	nox_swprintf(dst, format, amount);
}
*/
import "C"

import "unsafe"

func shopString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(shopSource), line)
}

func shopItemAttrs(dr *nox_drawable) *C.nox_item_amount_const_attrs_t {
	if dr.flags28&0x13001000 != 0 {
		return (*C.nox_item_amount_const_attrs_t)(unsafe.Pointer(drawableItemData(dr)))
	}
	return nil
}

//export sub_478730
func sub_478730(point *C.int) {
	pos := (*C.int2)(unsafe.Pointer(point))
	cell := shopCellAt(pos)
	if cell.count == 0 {
		return
	}
	gold, quantity := C.uint32_t(sub_4674A0()), cell.count
	if gold < quantity*cell.price {
		quantity = gold / cell.price
	}
	if quantity == 0 {
		sub_479520(C.int(cell.price - gold))
		return
	}
	dr := cell.drawable
	sub_4C05F0(1, C.int(cell.price))
	nox_gui_itemAmountDialog_4C0430(shopString("BuyLabel", 328), pos.field_0, pos.field_4,
		cell.net_codes[cell.count-1], dr.field_27, shopItemAttrs(dr), quantity, 0,
		(C.nox_gui_item_amount_callback_t)(C.sub_478850), nil)
}

func drawShopInstructions(pos *C.int2, text **wchar2_t, key string, line int32, white bool) C.int {
	inventory := asWindow(nox_xxx_wndGetChildByID_46B0C0(shopRoot(), 3806))
	origin, size := inventory.GlobalPos(), inventory.Size()
	if white {
		nox_xxx_drawSetTextColor_434390(int32(C.nox_color_white_2523948))
	}
	nox_client_drawImageAt_47D2C0(C.dword_5d4594_1098456, int32(pos.field_0), int32(pos.field_4))
	if *text == nil && key != "" {
		*text = shopString(key, line)
	}
	if *text == nil {
		return 0
	}
	return C.int(nox_xxx_drawStringWrap_43FAF0(nil, *text, int32(origin.X)+8, int32(origin.Y)+8,
		int32(size.X)-16, int32(size.Y)-16))
}

//export sub_478A70
func sub_478A70(pos *C.int2) C.int {
	return drawShopInstructions(pos, &C.dword_5d4594_1098604, "", 0, true)
}

//export sub_478B10
func sub_478B10(pos *C.int2) *wchar2_t {
	drawShopInstructions(pos, &C.dword_5d4594_1098596, "SellInstructions", 597, false)
	return C.dword_5d4594_1098596
}

//export sub_478BC0
func sub_478BC0(pos *C.int) *wchar2_t {
	drawShopInstructions((*C.int2)(unsafe.Pointer(pos)), &C.dword_5d4594_1098600, "RepairInstructions", 628, true)
	return C.dword_5d4594_1098600
}

//export nox_xxx_cliStartShopDlg_478FD0
func nox_xxx_cliStartShopDlg_478FD0(name *C.nox_shop_const_wchar2_t, intro *C.char, typeID C.int) C.int {
	root := shopRoot()
	label := nox_xxx_wndGetChildByID_46B0C0(root, 3810)
	sub_445C20()
	C.dword_5d4594_1098624, C.dword_5d4594_1098628 = 1, 1
	nox_window_set_hidden(root, 0)
	nox_xxx_wnd_46ABB0_go(root, 1)
	nox_xxx_wndShowModalMB_46A8C0(root)
	*g2ServerWord(1098612) = uint32(nox_client_getRenderGUI())
	nox_client_setRenderGUI(0)
	C.sub_467BB0()
	if name == nil {
		name = (*C.nox_shop_const_wchar2_t)(g2ServerMem(1107044))
	}
	title := (*wchar2_t)(g2ServerMem(1097300))
	C.nox_wcscpy(title, (*wchar2_t)(unsafe.Pointer(name)))
	C.sub_46AEE0(label, title)
	if intro != nil && *intro != 0 {
		C.dword_5d4594_1098604 = nox_strman_loadString_40F1D0(intro, &C.nox_shop_dialog_file_1098608, internCStr(shopSource), 1128)
	} else {
		C.dword_5d4594_1098604, C.nox_shop_dialog_file_1098608 = nil, nil
	}
	nox_xxx_getShopPic_4790F0(typeID)
	if file := C.nox_shop_dialog_file_1098608; file != nil {
		nox_xxx_playDialogFile_44D900((*byte)(unsafe.Pointer(file)), 100)
	}
	C.dword_5d4594_1107036 = 0
	return C.int(g2ServerEvent(C.dword_5d4594_1098580, 16394, uintptr(C.dword_5d4594_1098592), 0))
}

//export sub_479520
func sub_479520(amount C.int) {
	format := shopString("NotEnoughGold", 1346)
	text := (*wchar2_t)(g2ServerMem(1097352))
	C.shop_format_gold_shortage(text, format, amount)
	nox_xxx_dialogMsgBoxCreate_449A10(shopRoot(), shopString("ShopInformationTitle", 1350), text, 33, nil, nil)
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100)
}

//export sub_4795E0
func sub_4795E0(code, price C.int) C.int {
	pos := nox_client_getMousePos_4309F0()
	active := C.dword_5d4594_1098616
	if active == 1 {
		return 1
	}
	dr := C.sub_4676D0(code)
	if dr == nil {
		return C.int(active)
	}
	sub_4C05F0(1, price)
	quantity := C.sub_467700(code)
	result := nox_gui_itemAmountDialog_4C0430(shopString("SellLabel", 1504), pos.x, pos.y,
		C.uint32_t(code), dr.field_27, shopItemAttrs(dr), C.uint32_t(quantity), 0,
		(C.nox_gui_item_amount_callback_t)(C.sub_479690), (C.nox_gui_item_amount_callback_t)(C.sub_479680))
	C.dword_5d4594_1098616 = 1
	return result
}

//export sub_479740
func sub_479740(code C.int, price C.uint) {
	pos := nox_client_getMousePos_4309F0()
	gold := C.uint(sub_4674A0())
	if C.dword_5d4594_1098620 == 1 {
		return
	}
	dr := C.sub_4676D0(code)
	if dr == nil {
		return
	}
	sub_4C05F0(1, C.int(price))
	if price > gold {
		sub_479520(C.int(price - gold))
		sub_467680()
		return
	}
	nox_gui_itemAmountDialog_4C0430(shopString("RepairLabel", 1580), pos.x, pos.y,
		C.uint32_t(code), dr.field_27, shopItemAttrs(dr), 1, 0,
		(C.nox_gui_item_amount_callback_t)(C.sub_479820), (C.nox_gui_item_amount_callback_t)(C.sub_479810))
	C.dword_5d4594_1098620 = 1
}

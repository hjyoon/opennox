package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guiinv.h"
#include "client__gui__guimeter.h"
#include "client__gui__window.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

var inventoryWindows struct {
	panel, weapon, identify, slider, weaponHUD  *nox_window
	journal, stats, scrollUp, scrollDown, close *nox_window
}

//export nox_xxx_wndCreateInventoryMB_465E00
func nox_xxx_wndCreateInventoryMB_465E00() *C.nox_window {
	nox_xxx_inventoryLoadImages_467050()
	nox_xxx_inventoryNameSignInit_4671E0()
	inventoryFont = nox_xxx_guiFontPtrByName_43F360(internCStr("small"))
	width, height := clientDisplayStorage.width, clientDisplayStorage.height
	*memmap.PtrT[[6]uint32](0x5D4594, 1049732) = [6]uint32{0, 0, uint32(width), uint32(height), 0, 0}
	*memmap.PtrUint32(0x5D4594, 1049764) = uint32(width)
	*memmap.PtrUint32(0x5D4594, 1049768) = uint32(height)
	root := nox_window_new(nil, 552, 0, 0, 563, 264, nil)
	inventoryState.root = root
	if root == nil {
		return nil
	}
	nox_window_set_all_funcs(root, nil, C.nox_xxx_movEax1Sub_4661C0, nil)
	hover := nox_window_new(root, 8, 0, 224, width, 40, nil)
	nox_window_set_all_funcs(hover, C.nox_xxx_XorEaxEaxSub_464BA0, C.nox_xxx_movEax1Sub_4661C0, C.nox_xxx_inventroryOnHovewerSub_4667E0)

	inventory := nox_window_new(root, 40, 0, 0, 563, 224, C.nox_window_init_proc(C.sub_466220))
	inventoryWindows.panel = inventory
	if inventory == nil {
		return nil
	}
	nox_window_set_all_funcs(inventory, C.sub_464BD0, C.nox_xxx_inventoryDrawAllMB_463430, C.sub_466620)
	inventory.draw_data.style |= 0x100
	overlay := nox_window_new(root, 40, 0, 0, 1, 1, nil)
	nox_window_set_all_funcs(overlay, C.sub_464BD0, C.nox_xxx_movEax1Sub_4661C0, nil)
	weapon := nox_window_new(inventory, 40, 173, 174, 50, 50, nil)
	inventoryWindows.weapon = weapon
	nox_window_set_all_funcs(weapon, C.sub_464770, C.sub_4625D0, C.sub_4661D0)
	weapon.draw_data.style |= 0x100

	if sub_466950(inventory) == 0 || sub_466C40(inventory) == 0 || sub_466ED0(inventory) == 0 {
		return nil
	}
	hud := nox_window_new(nil, 0x408|C.NOX_WIN_LAYER_BACK, -1, height-127, 111, 127, nil)
	inventoryWindows.weaponHUD = hud
	if hud == nil {
		return nil
	}
	nox_window_set_all_funcs(hud, C.nox_xxx_inventoryWndProc_464BB0, C.nox_xxx_inventoryDrawProc_466580, nil)
	nox_xxx_wndSetIcon_46AE60(hud, nox_xxx_gLoadImg_42F970(internCStr("CurrentWeapon")))
	nox_xxx_wndSetIconLit_46AEA0(hud, nox_xxx_gLoadImg_42F970(internCStr("CurrentWeaponLit")))
	nox_xxx_wndSetOffsetMB_46AE40(hud, -1, 0)
	nox_win_init_cur_weapon(hud, 24, 51, 53, 53)
	sub_471160(hud, 79, 40, 20, 127)
	sub_470D70()
	toggle := nox_window_new(hud, 8, 5, 11, 28, 29, nil)
	nox_window_set_all_funcs(toggle, C.sub_466550, C.nox_xxx_movEax1Sub_4661C0, C.sub_466160)
	clientInventoryInitializeReservedCells()
	return inventory
}

func clientInventoryInitializeReservedCells() {
	for column := range inventoryColumns {
		for row := range inventoryRows {
			*inventoryGridCell(row, column) = C.nox_inventory_cell_t{}
		}
	}
	if legacyAddressableWords.dword_5d4594_1062560 == 0 {
		legacyAddressableWords.dword_5d4594_1062560 = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("Gold")))
		*memmap.PtrUint32(0x5D4594, 1049728) = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("QuestGoldPile")))
		*memmap.PtrUint32(0x5D4594, 1049724) = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("QuestGoldChest")))
	}
	entries := []struct {
		id   *C.uint32_t
		name string
	}{
		{&legacyAddressableWords.dword_5d4594_1062560, "Gold"},
		{&legacyAddressableWords.dword_5d4594_1062564, "Identify"},
		{&legacyAddressableWords.dword_5d4594_1062556, "AutoMap"},
	}
	for column, entry := range entries {
		if *entry.id == 0 && column != 0 {
			*entry.id = C.uint32_t(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(entry.name)))
		}
		cell := inventoryGridCell(inventoryRows-1, column)
		cell.field_0 = nox_new_drawable_for_thing(int32(*entry.id))
		if cell.field_0 != nil {
			cell.field_140 = 1
		}
	}
}

//export sub_466220
func sub_466220(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.uintptr_t {
	scroll := inventoryWindows.slider
	switch event {
	case 16393:
		if slider := game3Slider(scroll); slider != nil {
			legacyWords.dword_5d4594_1062512 = C.uint32_t(uintptr(slider.Max) - uintptr(arg2))
		}
	case 16391:
		control := (*nox_window)(unsafe.Pointer(uintptr(arg1)))
		id := nox_xxx_wndGetID_46B0A0(control)
		switch id {
		case 9102, 9103:
			slider := game3Slider(scroll)
			if slider == nil {
				return 0
			}
			value := int32(legacyWords.dword_5d4594_1062512)
			if id == 9102 {
				value = max(0, value-25) / 50 * 50
			} else {
				value += 50
				if value <= slider.Max {
					value -= value % 50
				} else {
					value = slider.Max
				}
			}
			legacyWords.dword_5d4594_1062512 = C.uint32_t(value)
			clientInventorySyncScroll(scroll, slider)
			nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		case 9105:
			height := max(0, sub_469FA0()-150)
			if inventoryState.mode == 5 {
				return 0
			}
			*memmap.PtrUint8(0x5D4594, 1049869) = 1
			legacyWords.dword_5d4594_1062516 = legacyWords.dword_5d4594_1062512
			legacyWords.dword_5d4594_1062512 = legacyWords.dword_5d4594_1062520
			clientInventoryChangeScrollRange(scroll, uintptr(height))
			clientInventorySetButton(inventoryWindows.journal, inventoryImages[C.NOX_INV_IMG_INVENTORY], inventoryImages[C.NOX_INV_IMG_INVENTORY_LIT], 9106)
		case 9106:
			*memmap.PtrUint8(0x5D4594, 1049869) = 0
			legacyWords.dword_5d4594_1062520 = legacyWords.dword_5d4594_1062512
			legacyWords.dword_5d4594_1062512 = legacyWords.dword_5d4594_1062516
			clientInventoryChangeScrollRange(scroll, 850)
			clientInventorySetButton(inventoryWindows.journal, nil, inventoryImages[C.NOX_INV_IMG_JOURNAL_LIT], 9105)
		case 9107:
			if inventoryState.mode != 5 {
				*memmap.PtrUint8(0x5D4594, 1049870) = 1
				clientInventorySetButton(inventoryWindows.stats, nil, inventoryImages[C.NOX_INV_IMG_DOLL_LIT], 9108)
				nox_window_set_hidden(inventoryWindows.weapon, 1)
			}
		case 9108:
			if inventoryState.mode != 5 {
				*memmap.PtrUint8(0x5D4594, 1049870) = 0
				clientInventorySetButton(inventoryWindows.stats, inventoryImages[C.NOX_INV_IMG_STATS], inventoryImages[C.NOX_INV_IMG_STATS_LIT], 9107)
				nox_window_set_hidden(inventoryWindows.weapon, 0)
			}
		case 9111:
			sub_467C10()
		}
	}
	return 0
}

func clientInventoryChangeScrollRange(win *nox_window, height uintptr) {
	slider := game3Slider(win)
	if slider == nil {
		return
	}
	nox_window_call_field_94_fnc(win, 16395, 0, height, internCStr("client_inventory_window_native.go"), 0)
	clientInventorySyncScroll(win, slider)
}

func clientInventorySyncScroll(win *nox_window, slider *game3SliderData) {
	value := uint32(slider.Max) - uint32(legacyWords.dword_5d4594_1062512)
	nox_window_call_field_94_fnc(win, 16394, uintptr(value), 0, internCStr("client_inventory_window_native.go"), 0)
}

func clientInventorySetButton(win *nox_window, image, selected *nox_video_bag_image_t, id int32) {
	nox_xxx_wndSetIcon_46AE60(win, image)
	sub_46AEC0(win, selected)
	nox_xxx_wndSetID_46B080(win, id)
}

//export sub_466550
func sub_466550(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int {
	switch event {
	case 5, 6:
		return 1
	case 7:
		nox_client_toggleInventory_467C60()
		return 1
	}
	return 0
}

//export sub_466950
func sub_466950(parent *C.nox_window) C.int {
	draw := C.nox_window_data{style: 1, win: parent, text_color: 0x80000000}
	for _, item := range []struct {
		out **C.nox_window
		y   C.int
		id  int32
		img int
		lit int
	}{
		{&inventoryWindows.scrollUp, 2, 9102, C.NOX_INV_IMG_UP, C.NOX_INV_IMG_UP_LIT},
		{&inventoryWindows.scrollDown, 148, 9103, C.NOX_INV_IMG_DOWN, C.NOX_INV_IMG_DOWN_LIT},
	} {
		draw.en_image = unsafe.Pointer(inventoryImages[item.img])
		draw.hl_image = unsafe.Pointer(inventoryImages[item.lit])
		draw.sel_image = draw.hl_image
		button := nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, 522, item.y, 20, 25, &draw)
		*item.out = button
		if button == nil {
			return 0
		}
		nox_xxx_wndSetID_46B080(button, item.id)
	}
	draw = C.nox_window_data{style: 8, win: parent, bg_color: 0x80000000, en_color: 0x80000000,
		hl_color: 0x80000000, dis_color: 0x80000000, sel_color: 0x80000000}
	values := game3SliderData{Max: 850}
	slider := nox_gui_newSlider_4B4EE0(parent, 1033, 524, 42, 16, 91, &draw, (*C.float)(unsafe.Pointer(&values)))
	inventoryWindows.slider = slider
	if slider == nil {
		return 0
	}
	nox_xxx_wndSetWindowProc_46B300(slider, C.sub_466BF0)
	nox_xxx_wndSetWindowProc_46B300(slider.field_100, C.sub_466BA0)
	slider.field_100.width, slider.field_100.height = 16, 16
	nox_xxx_wndSetOffsetMB_46AE40(slider.field_100, 0, -15)
	thumb := unsafe.Pointer(inventoryImages[C.NOX_INV_IMG_SLIDER])
	lit := unsafe.Pointer(inventoryImages[C.NOX_INV_IMG_SLIDER_LIT])
	sub_4B5700(slider, nil, nil, thumb, lit, lit)
	return 1
}

//export sub_466C40
func sub_466C40(parent *C.nox_window) C.int {
	for _, item := range []struct {
		out     **C.nox_window
		x, y, s C.int
		id      int32
		normal  *nox_video_bag_image_t
		lit     int
	}{
		{&inventoryWindows.journal, 243, 170, 34, 9105, nil, C.NOX_INV_IMG_JOURNAL_LIT},
		{&inventoryWindows.stats, 5, 186, 34, 9107, inventoryImages[C.NOX_INV_IMG_STATS], C.NOX_INV_IMG_STATS_LIT},
		{&inventoryWindows.close, 547, 2, 16, 9111, nil, C.NOX_INV_IMG_CLOSE_LIT},
	} {
		draw := C.nox_window_data{style: 1, win: parent, bg_image: unsafe.Pointer(item.normal),
			sel_image: unsafe.Pointer(inventoryImages[item.lit]), img_px: -C.intptr_t(item.x), img_py: -C.intptr_t(item.y)}
		button := nox_gui_newButtonOrCheckbox_4A91A0(parent, 1161, item.x, item.y, item.s, item.s, &draw)
		*item.out = button
		if button == nil {
			return 0
		}
		nox_xxx_wndSetID_46B080(button, item.id)
		nox_gui_winSetFunc96_46B070(button, C.sub_466E20)
	}
	return 1
}

//export sub_466ED0
func sub_466ED0(parent *C.nox_window) C.int {
	win := nox_new_window_from_file(internCStr("identify.wnd"), nil)
	inventoryWindows.identify = win
	if win == nil {
		return 0
	}
	C.sub_46AB20(win, 200, 200)
	sub_46B120(win, parent)
	nox_window_setPos_46A9B0(win, 51, 15)
	child := nox_xxx_wndGetChildByID_46B0C0(win, 9155)
	nox_xxx_wndSetDrawFn_46B340(child, C.sub_466F50)
	return 1
}

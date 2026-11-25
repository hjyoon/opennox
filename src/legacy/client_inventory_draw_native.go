package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guiinv.h"

extern uintptr_t dword_5d4594_1062492;
extern uint32_t dword_5d4594_1062496, dword_5d4594_1049844, nox_color_white_2523948;
extern nox_drawable* dword_5d4594_1063116;
extern void* nox_inventory_font;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export sub_465D50_draw
func sub_465D50_draw(win *C.nox_window) C.int {
	if win == nil || win.parent == nil {
		return 1
	}
	x, y := meterWindowPosition(win.parent)
	if item := sub_4615C0(); item != nil {
		item.pos.x, item.pos.y = C.intptr_t(x+51), C.intptr_t(y+81)
		if item.draw_func != nil {
			viewport := nox_client_inventory_viewport_native()
			ccall.CallIntPtr2(unsafe.Pointer(item.draw_func), unsafe.Pointer(viewport), unsafe.Pointer(item))
		}
	} else if C.dword_5d4594_1062496 == 0 && C.dword_5d4594_1062492 == 0 {
		nox_client_drawImageAt_47D2C0(C.nox_inventory_images[C.NOX_INV_IMG_FIST], int32(x+21), int32(y+50))
	}
	return 1
}

//export sub_465DE0
func sub_465DE0(value C.int) C.int {
	C.dword_5d4594_1049844 = C.uint32_t(value)
	return nox_xxx_j_inventoryNameSignInit_467460()
}

//export nox_xxx_movEax1Sub_4661C0
func nox_xxx_movEax1Sub_4661C0(win *C.nox_window, data unsafe.Pointer) C.int { return 1 }

//export nox_xxx_inventoryDrawProc_466580
func nox_xxx_inventoryDrawProc_466580(win *C.nox_window, draw *C.nox_window_data) C.int {
	x, y := meterWindowPosition(win)
	image := draw.bg_image
	if nox_client_inventory_animation_state() != 0 {
		image = draw.hl_image
	}
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(image), int32(x), int32(y))
	nox_xxx_drawSetTextColor_434390(int32(C.nox_color_white_2523948))
	if key := sub_42E8E0(35, 1); key != nil {
		nox_xxx_drawString_43F6E0(C.nox_inventory_font, key, int32(x+19), int32(y+102))
	}
	return 1
}

//export sub_466620
func sub_466620(control, _ C.int, packed C.uint) C.int {
	point := C.int2{field_0: C.int(uint16(packed)), field_4: C.int(packed >> 16)}
	nox_xxx_cursorSetTooltip_4776B0(sub_466660(control, &point))
	return 1
}

//export sub_466BA0
func sub_466BA0(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int {
	if nox_client_inventory_get_dragged() != nil {
		return C.sub_464BD0(win, event, arg1, arg2)
	}
	return nox_xxx_wndButtonProc_4A7F50(win, event, C.int(arg1), C.int(arg2))
}

//export sub_466BF0
func sub_466BF0(win *C.nox_window, event C.int, arg1, arg2 C.uintptr_t) C.int {
	if nox_client_inventory_get_dragged() != nil {
		return C.sub_464BD0(win, event, arg1, arg2)
	}
	return nox_xxx_wndScrollBoxDraw_4B4BA0(win, event, arg1, arg2)
}

//export sub_466F50
func sub_466F50(win *C.nox_window, draw *C.nox_window_data) C.int {
	item := C.dword_5d4594_1063116
	if item == nil {
		return 1
	}
	if item.flags28&0x13001000 != 0 {
		var definition unsafe.Pointer
		if item.flags28&0x11001000 != 0 {
			definition = nox_xxx_getProjectileClassById_413250(int32(item.field_27))
		} else {
			definition = nox_xxx_equipClothFindDefByTT_413270(int32(item.field_27))
		}
		applyEquipmentMaterials(definition, item)
	}
	x, y := meterWindowPosition(win)
	nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(draw.bg_image), int32(x)+int32(draw.img_px), int32(y)+int32(draw.img_py))
	return 1
}

//export nox_xxx_inventoryLoadImages_467050
func nox_xxx_inventoryLoadImages_467050() *C.nox_things_imageRef_t {
	names := [C.NOX_INV_IMG_COUNT]string{
		"InventoryBase", "InventoryIdentifyBase", "InventoryTray1", "InventoryTray2", "InventoryTray3",
		"InventoryTraySpecial", "InventoryTrayIdentifyLit", "InventoryTrayMapLit", "InventoryUpButton",
		"InventoryUpButtonLit", "InventoryDownButton", "InventoryDownButtonLit", "InventorySliderButton",
		"InventorySliderButtonLit", "InventoryEquipRing", "InventoryQuickItemRing", "InventoryCloseButtonLit",
		"InventoryJournalButtonLit", "InventoryInventoryButton", "InventoryInventoryButtonLit",
		"InventoryDollButtonLit", "InventoryStatsButton", "InventoryStatsButtonLit", "GUIFist", "SharedKeyMode",
	}
	for index, name := range names {
		C.nox_inventory_images[index] = nox_xxx_gLoadImg_42F970(internCStr(name))
	}
	C.nox_inventory_extra_lives_anim = nox_xxx_gLoadAnim_42FA20(internCStr("ExtraLives"))
	return C.nox_inventory_extra_lives_anim
}

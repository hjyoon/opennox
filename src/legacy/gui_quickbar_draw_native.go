package legacy

/*
#include "client__gui__guispell.h"
#include "noxstring.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func drawQuickbarSetName(win *nox_window, key string, line int32, set int, yOffset int32) {
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(win, &x, &y)
	var width, height C.int
	nox_window_get_size(win, &width, &height)
	var text [32]C.wchar2_t
	formatFixedWide(&text[0], quickbarText(key, line), formatValue{number: int32(set)})

	var textWidth C.int
	nox_xxx_drawGetStringSize_43F840(nil, &text[0], &textWidth, nil, 0)
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	nox_xxx_drawSetColor_4343E0(memmap.Int32(0x852978, 4))
	nox_draw_drawStringHL_43F730(nil, &text[0], int32(x)+int32(width-textWidth)/2, int32(y)+yOffset)
}

//export sub_45F9B0
func sub_45F9B0(win *C.nox_window, _ *C.nox_window_data) C.int {
	bar := g2qMain()
	if win == nil || bar == nil {
		return 0
	}
	if memmap.Uint32(0x5D4594, 1049476) == 0 {
		set := int(*g2qB(bar, 200)) + 1
		y := nox_xxx_guiFontHeightMB_43F320(quickbarControls.font) - nox_xxx_guiFontHeightMB_43F320(nil) + 1
		drawQuickbarSetName(win, "SpellSet", 1276, set, int32(y))
	}
	return 1
}

//export nox_xxx_quickbarDraw_45FAC0
func nox_xxx_quickbarDraw_45FAC0(win *C.nox_window, _ *C.nox_window_data) C.int {
	if win == nil {
		return 0
	}
	slot := int(memmap.Uint8(0x5D4594, 1048140)) + 1
	drawQuickbarSetName(win, "TrapSet", 1299, slot, 0)
	return 1
}

func quickbarButtonIndex(bar unsafe.Pointer, win *nox_window) int {
	for i := range 5 {
		if nox_quickbar_button(bar, C.int(i)) == win {
			return i
		}
	}
	return -1
}

func quickbarHighlighted(slot int) bool {
	return gameFrame() > 10 && uint32(slot) == memmap.Uint32(0x587000, 133484) &&
		gameFrame()-memmap.Uint32(0x5D4594, 1049540) < 10
}

func quickbarButtonIcon(icon *nox_video_bag_image_t, line int32, x, y int32) {
	if icon != nil {
		nox_client_drawImageAt_47D2C0(icon, x, y)
		return
	}
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	text := quickbarText("NoIcon", line)
	nox_xxx_drawString_43F6E0(nil, text, x+6, y+int32(nox_xxx_guiFontHeightMB_43F320(nil))+2)
}

//export nox_xxx_quickBarDrawFn_45FBD0
func nox_xxx_quickBarDrawFn_45FBD0(win *C.nox_window, draw *C.nox_window_data) C.int {
	if win == nil {
		return 0
	}
	bar := nox_quickbar_data_for_window(win)
	if bar == nil {
		return 0
	}
	slot := quickbarButtonIndex(bar, win)
	row := unsafe.Pointer(nox_quickbar_selected_row(bar))
	if slot < 0 || row == nil {
		return 0
	}
	ability := *g2qW(row, 8*uintptr(slot))
	if ability == 0 {
		return 1
	}
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(win, &x, &y)
	highlight := quickbarHighlighted(slot)
	icon := nox_xxx_spellGetAbilityIcon_425310(int32(ability), int32(bool2int(highlight)))
	title := nox_xxx_abilityGetName_0_425260(int32(ability))
	if draw == nil {
		draw = &win.draw_data
	}
	nox_xxx_wndWddSetTooltip_46B000(draw, title)
	quickbarButtonIcon(icon, 1388, int32(x), int32(y))

	state := uintptr(1047764 + 24*ability)
	if memmap.Uint32(0x5D4594, state+12) != 0 || memmap.Uint32(0x5D4594, state+8) == 0 ||
		nox_xxx_playerAnimCheck_4372B0() != 0 {
		cooldown := uint32(nox_xxx_abilityCooldown_4252D0(int32(ability)))
		step := cooldown / gameFPS()
		if step != 0 && !highlight {
			passed := (gameFrame() - memmap.Uint32(0x5D4594, state+20)) / step
			nox_client_drawRectFilledAlpha_49CF10(int32(x), int32(y), 34, int32(34-passed))
		}
	}
	return 1
}

//export nox_xxx_quickBarWarriorDraw_45FDE0
func nox_xxx_quickBarWarriorDraw_45FDE0(win *C.nox_window, draw *C.nox_window_data) C.int {
	if win == nil {
		return 0
	}
	bar := nox_quickbar_data_for_window(win)
	row := unsafe.Pointer(nox_quickbar_selected_row(bar))
	if row == nil {
		return 0
	}
	slot := quickbarButtonIndex(bar, win)
	if slot < 0 {
		return 0
	}
	spell := int32(*g2qW(row, 8*uintptr(slot)))
	if draw == nil {
		draw = &win.draw_data
	}
	if spell == 0 {
		empty := (*C.wchar2_t)(memmap.PtrOff(0x5D4594, 1049716))
		nox_xxx_wndWddSetTooltip_46B000(draw, empty)
		return 1
	}
	var x, y C.uint
	nox_client_wndGetPosition_46AA60(win, &x, &y)
	var icon unsafe.Pointer
	if quickbarHighlighted(slot) {
		icon = nox_xxx_spellIconHighlight_424AB0(spell)
	} else {
		icon = nox_xxx_spellIcon_424A90(spell)
	}
	flash := memmap.PtrUint8(0x5D4594, uintptr(1049544+spell))
	if int8(*flash) > 0 {
		*flash--
	}
	nox_xxx_wndWddSetTooltip_46B000(draw, nox_xxx_spellTitle_424930(spell))
	quickbarButtonIcon((*nox_video_bag_image_t)(icon), 1491, int32(x), int32(y))

	mana := int32(nox_xxx_cliGetMana_470DD0())
	cost := nox_xxx_spellManaCost_4249A0(spell, 1)
	if bar == g2qTrap() {
		for i := range slot {
			if prior := int32(*g2qW(row, 8*uintptr(i))); prior != 0 {
				mana -= nox_xxx_spellManaCost_4249A0(prior, 2)
			}
		}
	}
	if !bool(nox_xxx_spellIsEnabled_424B70(spell)) || nox_xxx_playerAnimCheck_4372B0() != 0 ||
		bool(nox_client_drawable_testBuff_4356C0((*nox_drawable)(*memmap.PtrPtr(0x852978, 8)), 29)) {
		nox_client_drawRectFilledAlpha_49CF10(int32(x), int32(y), 30, 30)
	} else if cost != 0 && mana < cost {
		nox_client_drawRectFilledAlpha_49CF10(int32(x), int32(y), 30, 30*(cost-mana)/cost)
	}
	return 1
}

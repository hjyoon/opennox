package legacy

/*
#include "client__gui__guispell.h"

extern nox_window* dword_5d4594_1049500;
extern nox_window* dword_5d4594_1049504;
extern nox_window* dword_5d4594_1049520;
extern uint32_t dword_5d4594_1049536;
extern int nox_win_height;
*/
import "C"

import (
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/common/memmap"
)

const quickbarStringsFile = "C:\\NoxPost\\src\\Client\\Gui\\guispell.c"

func quickbarText(key string, line int32) *C.wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(quickbarStringsFile), line)
}

func quickbarTrapError(key string, line int32) {
	nox_xxx_printCentered_445490(quickbarText(key, line))
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100)
}

//export nox_xxx_spellPutInBox_45DEB0
func nox_xxx_spellPutInBox_45DEB0(data unsafe.Pointer, spell, x, y C.int) C.int {
	slot := nox_xxx_spellBoxPointToWnd_45DE60(data, x, y)
	if slot < 0 {
		return 0
	}
	if data == g2qTrap() {
		if !bool(nox_xxx_spellCanUseInTrap_424BF0(int32(spell))) {
			quickbarTrapError("RestrictedTrapSpell", 496)
			return 0
		}
		row := unsafe.Pointer(nox_quickbar_selected_row(data))
		if row == nil {
			return 0
		}
		for i := range 3 {
			if *g2qW(row, 8*uintptr(i)) == uint32(spell) {
				quickbarTrapError("OneSpellPerTrap", 504)
				return 0
			}
		}
	}
	nox_xxx_spellKeyPackSetSpell_45DC40(data, spell, slot)
	return 1
}

//export nox_client_buildTrap_45E040
func nox_client_buildTrap_45E040() {
	row := unsafe.Pointer(nox_quickbar_trap_selected_row())
	if row == nil {
		return
	}
	var spells [4]uint32
	count := 0
	for i := range 3 {
		if id := *g2qW(row, 8*uintptr(i)); id != 0 {
			spells[count] = id
			count++
		}
	}
	const trapSpell = 34
	if count == 0 {
		if player := g2qPlayer(); player != nil && player.spell_lvl[trapSpell] != 0 {
			quickbarTrapError("TrapError", 1145)
		}
		return
	}
	*memmap.PtrUint8(0x5D4594, 1049488) = 0
	spells[count] = trapSpell
	flag := C.char(memmap.Uint8(0x5D4594, 1047924))
	nox_xxx_clientSendSpell_45DB20((*C.char)(unsafe.Pointer(&spells[0])), C.int(count+1), flag)
	*memmap.PtrUint32(0x5D4594, 1047916) = 0
	*memmap.PtrUint32(0x5D4594, 1049480) = 0
}

//export nox_xxx_quickbarButtonBook_45F3F0
func nox_xxx_quickbarButtonBook_45F3F0() C.int {
	key, line := "OpenSpellbookTT", int32(905)
	if sub_45CFC0() != 0 {
		key, line = "CloseSpellbookTT", 901
	}
	nox_xxx_cursorSetTooltip_4776B0(quickbarText(key, line))
	return 1
}

//export sub_45F480
func sub_45F480(win *C.nox_window, unusedDraw *C.nox_window_data, unusedEvent C.int) C.int {
	if win == nil {
		return 0
	}
	row, slot := uint32(win.field_92)>>16, uint32(win.field_92)&0xffff
	if row >= 5 || slot >= 5 {
		return 0
	}
	key, line := "ToolTipCastAtOther", int32(945)
	if sub_45F500(C.int(slot), g2qExpanded(int(row))) != 0 {
		key, line = "ToolTipCastOnMe", 941
	}
	nox_xxx_cursorSetTooltip_4776B0(quickbarText(key, line))
	return 1
}

//export sub_460070
func sub_460070() C.int {
	player := g2qPlayer()
	if player == nil {
		return 0
	}
	icon := C.dword_5d4594_1049504
	activate := C.dword_5d4594_1049520
	construct := C.dword_5d4594_1049500
	if icon == nil || activate == nil || construct == nil {
		return 1
	}
	var asset, tooltip string
	var line int32
	switch asPlayerS(player).PlayerClass() {
	case playerlib.Wizard:
		asset, tooltip, line = "QuickBarTrap", "ToolTipLayTrap", 1544
	case playerlib.Conjurer:
		asset, tooltip, line = "QuickBarBomber", "ToolTipSummonBomber", 1559
	default:
		return 1
	}
	nox_xxx_wndSetIcon_46AE60(icon, nox_xxx_gLoadImg_42F970(internCStr(asset)))
	nox_xxx_wndSetIconLit_46AEA0(icon, nox_xxx_gLoadImg_42F970(internCStr(asset+"Hit")))
	nox_xxx_wndWddSetTooltip_46B000(&activate.draw_data, quickbarText(tooltip, line))
	nox_xxx_wnd_46AD60(activate, 8)
	nox_xxx_wndWddSetTooltip_46B000(&construct.draw_data, quickbarText("ToolTipTrapConstruct", line+3))
	nox_xxx_wnd_46AD60(construct, 8)
	nox_xxx_wndSetDrawFn_46B340(construct, nil)
	return 1
}

//export nox_xxx_quickbarAddTrap_460EC0
func nox_xxx_quickbarAddTrap_460EC0(show C.int) *C.uint32_t {
	if show == 0 {
		sub_460070()
		return nil
	}
	construct := C.dword_5d4594_1049500
	icon := C.dword_5d4594_1049504
	if construct == nil || icon == nil || construct.flags&8 != 0 {
		return (*C.uint32_t)(unsafe.Pointer(construct))
	}
	C.dword_5d4594_1049536 = C.uint32_t(C.nox_win_height + 1)
	source := internCStr(quickbarStringsFile)
	var particle *C.nox_screenParticle
	for range 50 {
		interval := nox_common_randomIntMinMax_415FF0(4, 6, source, 2732)
		radius := nox_common_randomIntMinMax_415FF0(3, 6, source, 2731)
		vy := nox_common_randomIntMinMax_415FF0(-20, -5, source, 2729)
		vx := nox_common_randomIntMinMax_415FF0(-5, 5, source, 2728)
		y := nox_common_randomIntMinMax_415FF0(0, 20, source, 2727) + int32(icon.off_y) + 10
		x := nox_common_randomIntMinMax_415FF0(0, 20, source, 2726) + int32(icon.off_x) + 10
		particle = nox_client_newScreenParticle_431540(0, C.int(x), C.int(y), C.int(vx), C.int(vy), 1, C.char(radius), C.char(interval), 2, 1)
	}
	return (*C.uint32_t)(unsafe.Pointer(particle))
}

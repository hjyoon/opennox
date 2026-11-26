package legacy

/*
#include "GAME2_1.h"
#include "client__gui__guiinv.h"

*/
import "C"

import (
	"math"
	"strconv"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export nox_xxx_inventoryDrawAllMB_463430
func nox_xxx_inventoryDrawAllMB_463430(win *C.nox_window, draw *C.nox_window_data) C.int {
	nox_window_setPos_46A9B0(win.parent, 0, int32(legacyWords.dword_587000_136184))
	x, y := meterWindowPosition(win)
	nox_xxx_guiFontHeightMB_43F320(nil)
	clientInventoryDrawStatusIcons(int32(x)+10, int32(y)+234)
	if nox_client_inventory_animation_state() != 0 {
		if y+163 > 0 {
			left, top := x+254, y+13
			nox_xxx_wndDraw_49F7F0()
			currentDrawClipRect(left, top, 260, 150)
			switch memmap.Uint8(0x5D4594, 1049869) {
			case 0:
				nox_xxx_guiDrawInventoryTray_4643B0(left, top)
			case 1:
				nox_xxx_guiDrawJournal_469D40(left, top, C.int(legacyWords.dword_5d4594_1062512))
			}
			sub_49F860()
		}
		origin := [2]C.uint32_t{C.uint32_t(x), C.uint32_t(y)}
		if inventoryState.mode == 5 {
			sub_4627F0(&origin[0])
			nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_IDENTIFY_BASE], int32(x), int32(y))
		} else {
			switch memmap.Uint8(0x5D4594, 1049870) {
			case 0:
				sub_4BF7E0(&origin[0])
				nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_BASE], int32(x), int32(y))
			case 1:
				nox_client_makePlayerStatsDlg_463880((*C.int)(unsafe.Pointer(&origin[0])))
				nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_IDENTIFY_BASE], int32(x), int32(y))
			}
			nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
			text := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1062588))
			nox_xxx_drawStringWrap_43FAF0(nil, text, int32(x)+13, int32(y)+17, 196, 0)
		}
	}
	clientInventoryAdvanceAnimation()
	if sub_467C80() != 0 && nox_xxx_playerAnimCheck_4372B0() != 0 {
		sub_467C10()
	}
	return 1
}

func clientInventoryDrawStatusIcons(x, y int32) {
	for i := range 30 {
		if memmap.Uint32(0x5D4594, 1062540)&(uint32(1)<<i) != 0 {
			icon := nox_xxx_spellIcon_424A90(nox_xxx_getEnchantSpell_424920(int32(i)))
			nox_client_drawImageAt_47D2C0((*nox_video_bag_image_t)(icon), x, y)
			x += 35
		}
	}
	for i := range 6 {
		mask := byte(1) << i
		if memmap.Uint8(0x5D4594, 1062536)&mask != 0 {
			nox_client_drawImageAt_47D2C0(sub_413420(C.char(mask)), x, y)
			x += 35
		}
	}
	if noxflags.HasGame(noxflags.GameModeQuest) && inventoryExtraLives != nil {
		x, y = x+6, y+5
		anim := (*C.nox_image_ref_anim_t)(inventoryExtraLives.field_24)
		if anim != nil && anim.images_sz != 0 {
			images := unsafe.Slice(anim.images, int(anim.images_sz))
			nox_client_drawImageAt_47D2C0(images[gameFrame()%uint32(anim.images_sz)], x-58, y-53)
		}
		var buffer [16]uint16
		text := clientInventoryNumberText(&buffer, int32(memmap.Uint32(0x5D4594, 1050012)), "X ")
		nox_xxx_drawSetTextColor_434390(memmap.Int32(0x852978, 0))
		nox_xxx_drawString_43F6E0(inventoryFont, text, x+20, y+9)
		*memmap.PtrT[[4]int32](0x5D4594, 1049812) = [4]int32{x - 30, y - 20, x + 30, y + 20}
	}
	if noxflags.HasGame(noxflags.GameModeQuest) && sub_4BFD30() != 0 {
		x, y = x+66, y+5
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_SHARED_KEY_MODE], x-64, y-58)
		*memmap.PtrT[[4]int32](0x5D4594, 1049828) = [4]int32{x - 30, y - 20, x + 30, y + 20}
	}
}

func clientInventoryAdvanceAnimation() {
	state := memmap.PtrUint8(0x5D4594, 1049868)
	offset := int32(legacyWords.dword_587000_136184)
	switch *state {
	case 1:
		offset += 64
		if offset > 0 {
			offset = 0
			*state = 2
		}
		legacyWords.dword_587000_136184 = C.uint32_t(offset)
	case 3:
		offset -= 32
		if offset > -225 {
			legacyWords.dword_587000_136184 = C.uint32_t(offset)
			return
		}
		legacyWords.dword_587000_136184 = ^C.uint32_t(224)
		*state = 0
		journal := memmap.PtrUint8(0x5D4594, 1049869)
		switch *journal {
		case 0:
			legacyWords.dword_5d4594_1062516 = legacyWords.dword_5d4594_1062512
		case 1:
			legacyWords.dword_5d4594_1062520 = legacyWords.dword_5d4594_1062512
		}
		*journal = 0
		legacyWords.dword_5d4594_1062512 = legacyWords.dword_5d4594_1062516
		clientInventoryChangeScrollRange(inventoryWindows.slider, 850)
		clientInventorySetButton(inventoryWindows.journal, nil, inventoryImages[C.NOX_INV_IMG_JOURNAL_LIT], 9105)
		*memmap.PtrUint8(0x5D4594, 1049870) = 0
		clientInventorySetButton(inventoryWindows.stats, inventoryImages[C.NOX_INV_IMG_STATS], inventoryImages[C.NOX_INV_IMG_STATS_LIT], 9107)
		nox_window_set_hidden(inventoryWindows.weapon, 0)
	}
}

func clientInventoryNumberText(buffer *[16]uint16, number int32, prefix string) *wchar2_t {
	var bytes [16]byte
	text := strconv.AppendInt(append(bytes[:0], prefix...), int64(number), 10)
	for i, ch := range text {
		buffer[i] = uint16(ch)
	}
	buffer[len(text)] = 0
	return (*wchar2_t)(&buffer[0])
}

//export nox_xxx_guiDrawInventoryTray_4643B0
func nox_xxx_guiDrawInventoryTray_4643B0(left, top C.int) C.int {
	x, y := int32(left), int32(top)
	nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_TRAY_SPECIAL], x, y)
	var buffer [16]uint16
	text := clientInventoryNumberText(&buffer, int32(clientUIWords.player_gold), "")
	nox_xxx_drawSetTextColor_434390(int32(renderPalette.yellow))
	var width C.int
	nox_xxx_drawGetStringSize_43F840(inventoryFont, text, &width, nil, 0)
	nox_xxx_drawString_43F6E0(inventoryFont, text, x-int32(width)+43, y+36)
	if inventoryState.mode == 5 {
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_TRAY_IDENTIFY_LIT], x, y+50)
	}
	if sub_473670() != 0 {
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_TRAY_MAP_LIT], x, y+100)
	}
	for row := range inventoryRows {
		rowY := y - int32(legacyWords.dword_5d4594_1062512) + 50*int32(row)
		if rowY <= y-50 {
			continue
		}
		if rowY > y+150 {
			break
		}
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_TRAY1+row%3], x+60, rowY)
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_140 != 0 && cell.field_0 != nil {
				clientInventoryDrawTrayCell(cell, x+60+50*int32(column), rowY, &buffer)
			}
		}
	}
	return top + 150
}

func clientInventoryDrawTrayCell(cell *C.nox_inventory_cell_t, x, y int32, buffer *[16]uint16) {
	drawable := cell.field_0
	nox_client_drawEnableAlpha_434560(1)
	nox_client_drawSetAlpha_434580(0x40)
	durability, maximum := float64(drawable.field_73_1), float64(drawable.field_73_2)
	color := uint32(0x80000000)
	if durability < maximum*memmap.Float64(0x581450, 9608) {
		color = memmap.Uint32(0x85B3FC, 940)
	} else if durability < maximum*math.Float64frombits(uint64(legacyWords.qword_581450_9544)) {
		color = uint32(renderPalette.yellow)
	}
	if color != 0x80000000 {
		nox_client_drawSetColor_434460(int32(color))
		nox_client_drawRectFilledOpaque_49CE30(x, y, 50, 50)
	}
	nox_client_drawSetAlpha_434580(0x80)
	switch {
	case cell.field_132 != 0:
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_EQUIP_RING], x, y)
	case cell.field_136 != 0:
		nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_QUICK_ITEM_RING], x, y)
	case cell.field_4 == clientItemStorage.equipped_item_code:
		if alt := clientInventoryAlternateCell(); alt != nil && alt.field_0 != nil &&
			alt.field_0.flags28&0x1000000 != 0 && alt.field_0.flags29&0xc != 0 {
			nox_client_drawImageAt_47D2C0(inventoryImages[C.NOX_INV_IMG_QUICK_ITEM_RING], x, y)
		}
	}
	nox_client_drawEnableAlpha_434560(0)
	drawable.pos.x, drawable.pos.y = C.intptr_t(x+25), C.intptr_t(y+25)
	if drawable.draw_func != nil {
		ccall.CallIntPtr2(unsafe.Pointer(drawable.draw_func), unsafe.Pointer(nox_client_inventory_viewport_native()), unsafe.Pointer(drawable))
	}
	item := drawableItemData(drawable)
	if inventoryState.mode == 6 {
		depletedAmmo := drawable.flags28&0x1000 != 0 && item.Field_112_2 != 0 && item.Field_112_0 < item.Field_112_2
		if (drawable.field_73_1 == drawable.field_73_2 || drawable.field_73_2 == 0) && !depletedAmmo {
			nox_client_drawRectFilledAlpha_49CF10(x, y, 50, 50)
		}
	}
	if cell.field_140 > 1 {
		text := clientInventoryNumberText(buffer, int32(cell.field_140), "")
		nox_xxx_drawSetTextColor_434390(int32(renderPalette.white))
		nox_xxx_drawString_43F6E0(inventoryFont, text, x+6, y+6)
	}
	if drawable.flags28&0x13001000 != 0 && item.Field_112_0 >= 0 {
		text := clientInventoryNumberText(buffer, int32(item.Field_112_0), "")
		nox_xxx_drawSetTextColor_434390(int32(renderPalette.blue))
		var width C.int
		nox_xxx_drawGetStringSize_43F840(inventoryFont, text, &width, nil, 0)
		nox_xxx_drawString_43F6E0(inventoryFont, text, x-int32(width)+44, y+6)
	}
}

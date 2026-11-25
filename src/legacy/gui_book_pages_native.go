package legacy

/*
#include "client__gui__guibook.h"
#include "noxstring.h"

extern uint32_t dword_5d4594_1046656;
extern uint32_t dword_5d4594_1046868;
extern uint32_t dword_5d4594_1046872;
extern uint32_t dword_5d4594_1046932;
extern uint32_t dword_5d4594_1046936;
extern uint32_t dword_5d4594_1047512;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern nox_things_imageRef_t* dword_5d4594_1046924;
extern nox_things_imageRef_t* dword_5d4594_1046928;
extern nox_window* dword_5d4594_1046944;
extern nox_window* dword_5d4594_1046948;
extern nox_video_bag_image_t* nox_book_background_image;
extern nox_video_bag_image_t* nox_book_guide_tab_lit_image;
extern nox_video_bag_image_t* nox_book_spell_tab_lit_image;

static inline void nox_book_format_level(wchar2_t* dst, const wchar2_t* format, int level) {
	nox_swprintf(dst, format, level);
}
*/
import "C"

import (
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

type bookPageLayout struct {
	x, y, fontHeight int32
}

func (page bookPageLayout) drawTitle(text *C.wchar2_t, y int32) int32 {
	var width, height C.int
	nox_xxx_bookGetStringSize_43FA80(nil, text, &width, &height, 108)
	x := page.x + 24
	if int32(height) <= page.fontHeight {
		x += (108 - int32(width)) / 2
	}
	nox_xxx_bookDrawString_43FA80_43FD80(nil, text, x, y, 128, 0)
	return int32(height)
}

func (page bookPageLayout) drawDetail(text *C.wchar2_t) {
	if text == nil {
		return
	}
	var height C.int
	nox_xxx_drawGetStringSize_43F840(nil, text, nil, &height, 92)
	y := min(page.y+17+(141-int32(height))/2, page.y+52)
	nox_xxx_drawStringWrap_43FAF0(nil, text, page.x+153, y, 92, 0)
}

func (page bookPageLayout) drawCenteredLine(text *C.wchar2_t, y int32) {
	var width C.int
	nox_xxx_drawGetStringSize_43F840(nil, text, &width, nil, 0)
	x := page.x + 24 + (108-int32(width))/2
	nox_xxx_drawStringWrap_43FAF0(nil, text, x, y, 128, 0)
}

func (page bookPageLayout) drawContents(player *C.nox_playerInfo) {
	if C.dword_5d4594_1046936 == 0 {
		nox_window_set_hidden(C.dword_5d4594_1046944, 1)
	}
	spacing := int32(C.dword_5d4594_1046656)
	rows := 141/spacing - 1
	x, y := page.x+78, page.y+19
	for slot := int32(0); slot < 2*rows; slot++ {
		index := slot + 2*rows*int32(C.dword_5d4594_1046936)
		if index >= memmap.Int32(0x5D4594, 1047508)-int32(C.dword_5d4594_1047512) {
			break
		}
		if slot == rows {
			x, y = page.x+199, page.y+19
		}
		nox_xxx_drawSetTextColor_434390(memmap.Int32(0x5D4594, 1046880))
		class := asPlayerS(player).PlayerClass()
		id := int32(game2BookID(uint32(index)))
		var title *C.wchar2_t
		switch {
		case C.dword_5d4594_1046868 == 1:
			if class == playerlib.Conjurer && !bool(nox_xxx_spellIsEnabled_424B70(id+74)) {
				nox_xxx_drawSetTextColor_434390(memmap.Int32(0x5D4594, 1046884))
			}
			title = nox_xxx_guiCreatureGetName_427240(C.int(id))
		case class != playerlib.Warrior:
			if !bool(nox_xxx_spellIsEnabled_424B70(id)) {
				nox_xxx_drawSetTextColor_434390(memmap.Int32(0x5D4594, 1046884))
			}
			title = nox_xxx_spellTitle_424930(id)
		default:
			title = nox_xxx_abilityGetName_0_425260(id)
		}
		var width C.int
		nox_xxx_drawGetStringSize_43F840(nil, title, &width, nil, 128)
		nox_xxx_drawString_43F6E0(nil, title, x-int32(width)/2, y)
		y += int32(C.dword_5d4594_1046656)
	}
}

func (page bookPageLayout) drawCreature(player *C.nox_playerInfo, id C.int) {
	var sizeName string
	var sourceLine int32
	width, gap := int32(38), int32(0)
	switch nox_xxx_guideGetUnitSize_427460(id) {
	case 1:
		sizeName, sourceLine, gap = "Small", 983, 19
	case 2:
		sizeName, sourceLine = "Medium", 989
	case 4:
		sizeName, sourceLine, width = "Large", 995, 76
	}
	var text [256]C.wchar2_t
	copyWideString(&text[0], bookText("Size", 978))
	appendWideString(&text[0], internWStr(" "))
	if sizeName != "" {
		appendWideString(&text[0], bookText(sizeName, sourceLine))
	}
	if asPlayerS(player).PlayerClass() == playerlib.Conjurer &&
		(player.spell_lvl[134] != 0 || noxflags.HasGame(noxflags.GameOnline) && !noxflags.HasGame(noxflags.GameModeQuest)) {
		page.drawCenteredLine(&text[0], page.y+51)
	}
	y := page.y + 53 + page.fontHeight
	height := page.drawTitle(nox_xxx_guiCreatureGetName_427240(id), y)
	icon := nox_xxx_bookGetCreatureImg_427400(id)
	nox_client_drawImageAt_47D2C0(icon, page.x+24+(108-width)/2, y+gap+height+2)
	page.drawDetail(nox_xxx_guideGetDescById_4272E0(id))
}

func (page bookPageLayout) drawSpell(player *C.nox_playerInfo, id int32) {
	flags := uint16(nox_xxx_spellFlags_424A70(id))
	y := page.y + 53
	y += page.drawTitle(nox_xxx_spellTitle_424930(id), y) + 4
	cost := nox_xxx_spellManaCost_4249A0(id, 1)
	var text [256]C.wchar2_t
	copyWideString(&text[0], bookText("ManaCost", 1115))
	appendWideString(&text[0], internWStr(" "))
	switch {
	case cost != 0:
		var number [12]C.wchar2_t
		nox_itow(C.int(cost), &number[0], 10)
		appendWideString(&text[0], &number[0])
	case bool(nox_xxx_spellHasFlags_424A50(id, 0x800000)):
		appendWideString(&text[0], internWStr("0"))
	default:
		appendWideString(&text[0], internWStr("*"))
	}
	page.drawCenteredLine(&text[0], y)
	y += page.fontHeight + 2
	for _, row := range [...]struct {
		flag uint16
		key  string
		line int32
	}{
		{0x100, "SpellInstant", 1173},
		{4, "SpellTargeted", 1181},
		{8, "SpellAtLocation", 1189},
		{0x20, "SpellHostile", 1197},
	} {
		if flags&row.flag != 0 {
			page.drawCenteredLine(bookText(row.key, row.line), y)
			y += page.fontHeight
		}
	}
	C.nox_book_format_level(&text[0], bookText("PowerLevel", 1206), C.int(player.spell_lvl[id]))
	page.drawCenteredLine(&text[0], y+page.fontHeight)
	page.drawDetail(nox_xxx_spellDescription_424A30(id))
}

//export nox_xxx_bookDrawList_45BD40
func nox_xxx_bookDrawList_45BD40(win *C.nox_window) C.int {
	player := bookSelectedPlayer()
	var x, y C.uint
	nox_gui_getWindowOffs_46AA20(win, &x, &y)
	page := bookPageLayout{int32(x), int32(y), nox_xxx_guiFontHeightMB_43F320(nil)}
	C.dword_5d4594_1046656 = C.uint32_t(page.fontHeight + 2)
	nox_xxx_drawSetTextColor_434390(memmap.Int32(0x5D4594, 1046880))
	nox_client_drawImageAt_47D2C0(C.nox_book_background_image, page.x-24, page.y-76)
	tab := C.nox_book_spell_tab_lit_image
	if C.dword_5d4594_1046872 != 0 {
		tab = C.nox_book_guide_tab_lit_image
	}
	nox_client_drawImageAt_47D2C0(tab, page.x-24, page.y-76)

	switch C.dword_5d4594_1046868 {
	case 3:
		nox_video_drawAnimatedImageOrCursorAt_4BE6D0(uintptr(unsafe.Pointer(C.dword_5d4594_1046928)), page.x-24, page.y-76)
	case 2:
		nox_video_drawAnimatedImageOrCursorAt_4BE6D0(uintptr(unsafe.Pointer(C.dword_5d4594_1046924)), page.x-24, page.y-76)
	default:
		if C.nox_xxx_aNox_cfg_0_587000_132132 != 0 {
			page.drawContents(player)
		} else {
			id := int32(game2BookID(uint32(C.dword_5d4594_1046932)))
			switch {
			case C.dword_5d4594_1046868 == 1:
				page.drawCreature(player, C.int(id))
			case asPlayerS(player).PlayerClass() == playerlib.Warrior:
				sub_425450(id)
				page.drawTitle(nox_xxx_abilityGetName_0_425260(id), page.y+53)
				page.drawDetail(sub_4252F0(id))
			default:
				page.drawSpell(player, id)
			}
		}
	}
	if C.nox_xxx_aNox_cfg_0_587000_132132 == 0 &&
		int32(C.dword_5d4594_1046932) >= memmap.Int32(0x5D4594, 1047508)-int32(C.dword_5d4594_1047512)-1 {
		nox_window_set_hidden(C.dword_5d4594_1046948, 1)
	}
	return 1
}

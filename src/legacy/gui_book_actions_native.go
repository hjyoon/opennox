package legacy

/*
#include "client__gui__guibook.h"
*/
import "C"

import noxflags "github.com/opennox/opennox/v1/common/flags"

const bookStringsFile = "C:\\NoxPost\\src\\Client\\Gui\\guibook.c"

func bookText(key string, line int32) *C.wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(bookStringsFile), line)
}

//export nox_xxx_bookShowMB_45AD70
func nox_xxx_bookShowMB_45AD70(force C.int) {
	player := g2qPlayer()
	if nox_xxx_guiCursor_477600() != 0 {
		return
	}
	if nox_xxx_playerAnimCheck_4372B0() != 0 && noxflags.HasGame(noxflags.GameModeCoop) {
		return
	}
	if player == nil || nox_xxx_guiSpellSortList_45ADF0(C.int(game2PlayerClass(player))) != 0 {
		nox_xxx_book_45B010(force)
		return
	}
	nox_xxx_printCentered_445490(bookText("EmptyBook", 377))
	nox_xxx_clientPlaySoundSpecial_452D80(925, 100)
}

//export nox_xxx_book_45CF00
func nox_xxx_book_45CF00(data *C.uint32_t) C.int {
	switch *data {
	case 1310:
		player := bookSelectedPlayer()
		key, line := "ToolTipSpellTab", int32(1644)
		if game2PlayerClass(player) == 0 {
			key, line = "ToolTipAbilityTab", 1640
		}
		nox_xxx_cursorSetTooltip_4776B0(bookText(key, line))
	case 1320:
		nox_xxx_cursorSetTooltip_4776B0(bookText("ToolTipGuideTab", 1649))
	}
	return 1
}

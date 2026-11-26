package legacy

/*
#include "GAME3.h"
#include "client__shell__selclass.h"
#include "client__shell__selcolor.h"

*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const classSelectionStringsFile = "C:\\NoxPost\\src\\client\\shell\\SelClass.c"

//export sub_4A4A20
func sub_4A4A20(unusedWindow *C.nox_window, event C.int, source *C.nox_window, unusedArg C.int) C.int {
	switch event {
	case 16389:
		id := nox_xxx_wndGetID_46B0A0(source)
		if id >= 601 && id <= 603 {
			index := id - 601
			g2ServerEnabled(classSelectNext, true)
			description := nox_xxx_wndGetChildByID_46B0C0(classSelectRoot, 605)
			*(*byte)(unsafe.Add(unsafe.Pointer(classSelectProfile), 66)) = byte(index)
			idPtr := (*C.char)(memmap.PtrOff(0x587000, 170220+8*uintptr(index)))
			text := nox_strman_loadString_40F1D0(idPtr, nil, internCStr(classSelectionStringsFile), 279)
			g2ServerText(description, 16385, unsafe.Pointer(text), 0)
			*memmap.PtrUint32(0x5D4594, 1307740) = uint32(id)
		}
		nox_xxx_clientPlaySoundSpecial_452D80(920, 100)
		return 1
	case 16391:
		id := nox_xxx_wndGetID_46B0A0(source)
		if id >= 601 && id <= 603 {
			return 1
		}
		if id == 610 {
			if noxflags.HasGame(noxflags.GameOnline) && !noxflags.HasGame(noxflags.GameModeQuest) {
				quest := nox_xxx_isQuest_4D6F50() != 0 || sub_4D6F70() != 0
				sub_4A4B70(C.int(bool2int(quest)))
			}
			sub_4A4970()
			classSelectAnimation.Func13Ptr = C.nox_game_showSelColor_4A5D00
		}
		nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
		return 1
	default:
		return 0
	}
}

//export sub_4A4B70
func sub_4A4B70(clearSpells C.int) C.uintptr_t {
	class := *(*byte)(unsafe.Add(unsafe.Pointer(classSelectProfile), 66))
	tables := [...]struct{ offset, length uintptr }{
		{170104, 12}, {170116, 20}, {170136, 20},
	}
	if int(class) >= len(tables) {
		return 0
	}
	table := tables[class]
	rows := memmap.Slice(0x587000, table.offset)[:table.length]
	count := 0
	for count*5+5 <= len(rows) && rows[5*count] != 0 {
		count++
	}
	if count == 0 {
		return C.uintptr_t(class)
	}
	row := nox_common_randomIntMinMax_415FF0(0, int32(count-1), internCStr(classSelectionStringsFile), 195)
	var spells [5]byte
	if clearSpells != 1 {
		copy(spells[:], rows[5*int(row):])
	}
	if class != 0 {
		for set := range 5 {
			nox_xxx_clientUpdateButtonRow_45E110(C.int(set))
			for slot, spell := range spells {
				nox_game2_book_insert_native(2, C.int(spell), C.int(slot))
			}
		}
		return C.uintptr_t(nox_xxx_clientUpdateButtonRow_45E110(0))
	}
	nox_xxx_clientUpdateButtonRow_45E110(0)
	var result C.uintptr_t
	for slot, ability := range spells {
		result = nox_game2_book_insert_native(3, C.int(ability), C.int(slot))
	}
	return result
}

package legacy

/*
#include "client__gui__guicon.h"
#include "client__gui__guimsg.h"
#include "common__system__settings.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const centerMessageWords = 318

//export nox_xxx_printCentered_445490
func nox_xxx_printCentered_445490(text *C.wchar2_t) {
	if text == nil {
		return
	}
	index := clientItemStorage.center_message_index + 1
	if index >= 3 {
		index = 0
	}
	clientItemStorage.center_message_index = index
	offset := 644 * uintptr(index)
	destination := (*[centerMessageWords]uint16)(game1CenterMemory(0x5D4594, 823804+offset))
	alloc.StrCopyZero16P(destination[:], (*uint16)(unsafe.Pointer(text)))
	*(*uint32)(game1CenterMemory(0x5D4594, 824440+offset)) = gameFrame() + 4*gameFPS() + gameFPS()
	*(*byte)(game1CenterMemory(0x5D4594, 824444+offset)) = 0
	format := nox_strman_loadString_40F1D0(internCStr("systemmsg"), nil,
		internCStr(`C:\NoxPost\src\Client\Gui\guimsg.c`), 69)
	consolePrintRed(format, formatValue{text: unsafe.Pointer(text)})
}

func centerMessageWithNumber(format *wchar2_t, value C.int) {
	var text [centerMessageWords]wchar2_t
	formatWideNumber(text[:], format, value)
	nox_xxx_printCentered_445490(&text[0])
}

func centerMessageWithName(format, name *wchar2_t) {
	var text [centerMessageWords]wchar2_t
	formatWideStrings(text[:], format, name, nil)
	nox_xxx_printCentered_445490(&text[0])
}

//export sub_409FB0_settings
func sub_409FB0_settings(mode C.short, limit C.ushort) {
	index := sub_409A70(mode)
	cell := memmap.PtrUint16(0x5D4594, 3488+2*uintptr(index))
	if *cell == uint16(limit) {
		return
	}
	limit = min(999, limit)
	*cell = uint16(limit)
	nox_server_gameSettingsUpdated_40A670()
	if nox_client_isConnected_43C700() != 0 {
		format := nox_strman_loadString_40F1D0(internCStr("parsecmd.c:lessonsset"), nil,
			internCStr(`C:\NoxPost\src\Common\System\settings.c`), 389)
		centerMessageWithNumber(format, C.int(limit))
	}
}

package legacy

/*
#include "GAME3_2.h"

typedef int (*gui_settings_fn)(char, const void*, int);
typedef void (*new_player_packet_fn)(unsigned char*, nox_playerInfo*);
_Static_assert(_Generic(&nox_xxx_netGuiGameSettings_4DD9B0, gui_settings_fn:1, default:0), "GUI settings ABI");
_Static_assert(_Generic(&nox_xxx_netNewPlayerMakePacket_4DDA90, new_player_packet_fn:1, default:0), "new player packet ABI");
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callGUISettingsPacketC(code byte, settings *byte, index int32) int32 {
	return int32(C.nox_xxx_netGuiGameSettings_4DD9B0(C.char(code), unsafe.Pointer(settings), C.int(index)))
}

func callNewPlayerPacketC(buf *byte, player *server.Player) {
	C.nox_xxx_netNewPlayerMakePacket_4DDA90((*C.uchar)(unsafe.Pointer(buf)), (*C.nox_playerInfo)(player.C()))
}

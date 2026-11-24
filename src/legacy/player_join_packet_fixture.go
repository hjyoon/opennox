package legacy

/*
#include "GAME3_2.h"

typedef int (*gui_settings_fn)(char, const void*, int);
typedef void (*new_player_packet_fn)(unsigned char*, nox_playerInfo*);
typedef char* (*join_roster_fn)(int);
_Static_assert(_Generic(&nox_xxx_netGuiGameSettings_4DD9B0, gui_settings_fn:1, default:0), "GUI settings ABI");
_Static_assert(_Generic(&nox_xxx_netNewPlayerMakePacket_4DDA90, new_player_packet_fn:1, default:0), "new player packet ABI");
_Static_assert(_Generic(&nox_xxx_newPlayerSendAllPlayers_4DE300, join_roster_fn:1, default:0), "join roster ABI");
extern uint32_t dword_5d4594_1563276;
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

func callJoinRosterC(index int32) unsafe.Pointer {
	return unsafe.Pointer(C.nox_xxx_newPlayerSendAllPlayers_4DE300(C.int(index)))
}

func callJoinInventoryC(index int32, player *server.Player) {
	C.sub_4DDE10(C.int(index), (*C.nox_playerInfo)(player.C()))
}

func swapJoinFlagTypeC(value uint32) uint32 {
	old := C.dword_5d4594_1563276
	C.dword_5d4594_1563276 = C.uint32_t(value)
	return uint32(old)
}

func callGauntletNoticeC(index int32) int32 {
	return int32(C.nox_xxx_netGauntlet_4D9E70(C.int(index)))
}

func callJoinMinimapRevealC(index int32) int32 {
	return int32(C.nox_xxx_servMinimapRevealFlag_4DE380(C.int(index)))
}

func callJoinObjectSyncC(index int32) {
	C.sub_4DE410(C.int(index))
}

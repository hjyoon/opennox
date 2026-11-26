package cfixture

/*
#include "defs.h"
#include "client_native_storage.h"
static uint32_t score_packet_stamp_48ea70(nox_playerInfo* p) { return p->field_2160; }
static void score_packet_set_stamp_48ea70(nox_playerInfo* p, uint32_t value) { p->field_2160 = value; }
static size_t score_packet_stamp_offset_48ea70(void) { return offsetof(nox_playerInfo, field_2160); }
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func ScorePacketStamp48EA70(p *server.Player) uint32 {
	return uint32(C.score_packet_stamp_48ea70((*C.nox_playerInfo)(unsafe.Pointer(p))))
}

func SetScorePacketStamp48EA70(p *server.Player, value uint32) {
	C.score_packet_set_stamp_48ea70((*C.nox_playerInfo)(unsafe.Pointer(p)), C.uint32_t(value))
}

func ScorePacketStampOffset48EA70() int {
	return int(C.score_packet_stamp_offset_48ea70())
}

func ScorePacketCenteredIndex48EA70() uint32 {
	return uint32(C.nox_client_item_storage().center_message_index)
}

func SetScorePacketCenteredIndex48EA70(value uint32) {
	C.nox_client_item_storage().center_message_index = C.uint32_t(value)
}

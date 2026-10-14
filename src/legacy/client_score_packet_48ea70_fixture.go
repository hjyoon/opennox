package legacy

/*
#include "defs.h"
extern uint32_t dword_5d4594_825736;
static uint32_t score_packet_stamp_48ea70(nox_playerInfo* p) { return p->field_2160; }
static void score_packet_set_stamp_48ea70(nox_playerInfo* p, uint32_t value) { p->field_2160 = value; }
static size_t score_packet_stamp_offset_48ea70(void) { return offsetof(nox_playerInfo, field_2160); }
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func scorePacketStamp48EA70(p *server.Player) uint32 {
	return uint32(C.score_packet_stamp_48ea70((*C.nox_playerInfo)(unsafe.Pointer(p))))
}

func setScorePacketStamp48EA70(p *server.Player, value uint32) {
	C.score_packet_set_stamp_48ea70((*C.nox_playerInfo)(unsafe.Pointer(p)), C.uint32_t(value))
}

func scorePacketStampOffset48EA70() int {
	return int(C.score_packet_stamp_offset_48ea70())
}

func scorePacketCenteredIndex48EA70() uint32 {
	return uint32(C.dword_5d4594_825736)
}

func setScorePacketCenteredIndex48EA70(value uint32) {
	C.dword_5d4594_825736 = C.uint32_t(value)
}

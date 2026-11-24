package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netReportTotalHealth_4D85C0
func nox_xxx_netReportTotalHealth_4D85C0(recipient C.int, unit *nox_object_t) C.int {
	packet, ok := server.CreatureTotalHealthPacket(asObjectS(unit))
	if !ok {
		return 0
	}
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportAcquireCreature_4D91A0
func nox_xxx_netReportAcquireCreature_4D91A0(recipient C.int, unit *nox_object_t) C.int {
	packet := server.ObjectTypePacket(netmsg.MSG_REPORT_ACQUIRE_CREATURE, asObjectS(unit))
	if noxflags.HasGame(noxflags.GameFlag28) {
		packet[4] |= 0x80
	}
	sendImportantPacketBytes(recipient, packet[:], true, true)
	return nox_xxx_netReportTotalHealth_4D85C0(recipient, unit)
}

//export nox_xxx_netMonitorCreature_4D9250
func nox_xxx_netMonitorCreature_4D9250(recipient C.int, unit *nox_object_t) C.int {
	packet := server.ObjectTypePacket(netmsg.MSG_REPORT_MONITOR_CREATURE, asObjectS(unit))
	sendImportantPacketBytes(recipient, packet[:], true, true)
	return nox_xxx_netReportTotalHealth_4D85C0(recipient, unit)
}

func creatureUnmonitorPacket(unit *server.Object) [3]byte {
	code := uint16(server.ObjectNetCode(unit))
	return [3]byte{byte(netmsg.MSG_REPORT_UNMONITOR_CREATURE), byte(code), byte(code >> 8)}
}

//export nox_xxx_netSendUnMonitorCrea_4D92A0
func nox_xxx_netSendUnMonitorCrea_4D92A0(recipient C.int, raw *C.uint32_t) C.int {
	packet := creatureUnmonitorPacket((*server.Object)(unsafe.Pointer(raw)))
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportStatsSpeed_4D9360
func nox_xxx_netReportStatsSpeed_4D9360(recipient C.int, unit *nox_object_t, status C.char, speedBits C.int) C.int {
	packet := [8]byte{0: byte(netmsg.MSG_REPORT_STAT_MODIFIER), 7: byte(status)}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(server.ObjectNetCode(asObjectS(unit))))
	binary.LittleEndian.PutUint32(packet[3:7], uint32(speedBits))
	return sendImportantPacketBytes(recipient, packet[:], true, false)
}

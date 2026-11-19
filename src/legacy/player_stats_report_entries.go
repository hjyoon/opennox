package legacy

/*
#include "GAME3_2.h"
#include "GAME5_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netReportHealthDelta_4D8760
func nox_xxx_netReportHealthDelta_4D8760(recipient C.int, code, delta C.short) C.short {
	return C.short(netReportHealthDelta4D8760(int(recipient), uint16(code), int16(delta)))
}

func netReportHealthDelta4D8760(recipient int, code uint16, delta int16) int16 {
	if delta >= 0 {
		return delta
	}
	packet := [5]byte{byte(netmsg.MSG_REPORT_HEALTH_DELTA)}
	binary.LittleEndian.PutUint16(packet[1:3], code)
	binary.LittleEndian.PutUint16(packet[3:5], uint16(delta))
	return int16(sendImportantPacketWrapperC(recipient, packet[:], nil, 1, importantPacketSequenceDisabled))
}

//export nox_xxx_itemReportHealth_4D87A0
func nox_xxx_itemReportHealth_4D87A0(recipient C.int, item *nox_object_t) C.intptr_t {
	return netReportItemHealth4D87A0(int(recipient), asObjectS(item))
}

func netReportItemHealth4D87A0(recipient int, item *server.Object) C.intptr_t {
	if item == nil {
		return 0
	}
	health := item.HealthData
	if health == nil || health.Max == 0 {
		return C.intptr_t(uintptr(unsafe.Pointer(health)))
	}
	packet := server.BuildShopItemHealthPacket4D87A0(item)
	binary.LittleEndian.PutUint16(packet[1:3], uint16(server.ObjectNetCode(item)))
	return C.intptr_t(sendImportantPacketWrapperC(recipient, packet[:], nil, 0, importantPacketSequenceEnabled))
}

//export nox_xxx_servSendStats_4D8990
func nox_xxx_servSendStats_4D8990(recipient C.int, unit *nox_object_t, level C.char) C.char {
	obj := asObjectS(unit)
	if obj == nil {
		return 0
	}
	packet, ok := server.BuildPlayerStatsReport4D8990(obj, server.ObjectNetCode)
	if !ok {
		return C.char(obj.ObjClass)
	}
	packet[13] = byte(level)
	return C.char(sendImportantPacketWrapperC(int(recipient), packet[:], nil, 1, importantPacketSequenceDisabled))
}

//export nox_xxx_netReportArmorVal_4D8A30
func nox_xxx_netReportArmorVal_4D8A30(recipient, bits C.int) C.int {
	packet := server.BuildArmorValueReport4D8A30(uint32(bits))
	return C.int(sendImportantPacketWrapperC(int(recipient), packet[:], nil, 1, importantPacketSequenceDisabled))
}

//export nox_xxx_itemDestroyed_4E1650
func nox_xxx_itemDestroyed_4E1650(recipient C.int, item *nox_object_t, before, after C.ushort) C.intptr_t {
	obj := asObjectS(item)
	if obj == nil || obj.HealthData == nil {
		return 0
	}
	bucket, report := itemHealthReportDecision4E1650(obj, uint16(before), uint16(after))
	if report {
		return netReportItemHealth4D87A0(int(recipient), obj)
	}
	return C.intptr_t(bucket)
}

func itemHealthReportDecision4E1650(item *server.Object, before, after uint16) (int, bool) {
	if noxflags.HasGame(noxflags.GameModeCoop) {
		return 0, true
	}
	oldBucket := itemHealthBucket57B190(before, item.HealthData.Max)
	newBucket := itemHealthBucket57B190(after, item.HealthData.Max)
	return newBucket, oldBucket != newBucket
}

//export sub_57B190
func sub_57B190(current, maximum C.ushort) C.int {
	return C.int(itemHealthBucket57B190(uint16(current), uint16(maximum)))
}

func itemHealthBucket57B190(current, maximum uint16) int {
	// The original compares against exact binary halves and quarters.
	switch {
	case maximum == 0:
		return 4
	case current == maximum:
		return 0
	case 2*uint32(current) >= uint32(maximum):
		return 1
	case 4*uint32(current) >= uint32(maximum):
		return 2
	default:
		return 3
	}
}

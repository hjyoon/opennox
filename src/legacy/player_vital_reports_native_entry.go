package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netReportStamina_4D8800
func nox_xxx_netReportStamina_4D8800(recipient C.int, unit *nox_object_t) C.intptr_t {
	result, sent := GetServer().S().NetReportStamina4D8800(int32(recipient), asObjectS(unit))
	return vitalReportResult(unit, result, sent)
}

//export sub_4D8840
func sub_4D8840(recipient C.int, unit *nox_object_t) C.int {
	if unit == nil {
		return 0
	}
	return C.int(GetServer().S().NetReportItemEnchantment4D8840(int(recipient), asObjectS(unit)))
}

//export sub_4D8870
func sub_4D8870(recipient C.int, unit *nox_object_t) C.intptr_t {
	result, sent := GetServer().S().NetReportGold4D8870(int32(recipient), asObjectS(unit))
	return vitalReportResult(unit, result, sent)
}

//export nox_xxx_netReportTotalMana_4D88C0
func nox_xxx_netReportTotalMana_4D88C0(recipient C.int, unit *nox_object_t) C.intptr_t {
	result, sent := netReportManaNative(GetServer().S(), int(recipient), asObjectS(unit), true)
	return vitalReportResult(unit, int32(result), sent)
}

//export nox_xxx_netReportMana_4D8930
func nox_xxx_netReportMana_4D8930(recipient C.int, unit *nox_object_t) C.intptr_t {
	result, sent := netReportManaNative(GetServer().S(), int(recipient), asObjectS(unit), false)
	return vitalReportResult(unit, int32(result), sent)
}

// vitalReportResult preserves the sender's signed result or the unchanged
// unit address when the original entry skips the report.
func vitalReportResult(unit *nox_object_t, result int32, sent bool) C.intptr_t {
	if sent {
		return C.intptr_t(result)
	}
	return C.intptr_t(uintptr(unsafe.Pointer(unit)))
}

func NetReportTotalManaNative4D88C0(s *server.Server, recipient uint8, unit *server.Object) {
	netReportManaNative(s, int(recipient), unit, true)
}

func NetReportManaNative4D8930(s *server.Server, recipient uint8, unit *server.Object) int {
	result, _ := netReportManaNative(s, int(recipient), unit, false)
	return result
}

func netReportManaNative(s *server.Server, recipient int, unit *server.Object, includeMaximum bool) (int, bool) {
	if unit == nil || unit.ObjClass&object.ClassPlayer == 0 || unit.UpdateData == nil {
		return 0, false
	}
	update := (*server.PlayerUpdateData)(unit.UpdateData)
	if update.Player.PlayerClass() == 0 {
		return 0, false
	}
	var packet [7]byte
	length := 5
	packet[0] = byte(netmsg.MSG_REPORT_MANA)
	binary.LittleEndian.PutUint16(packet[1:], uint16(s.GetUnitNetCode(unit)))
	binary.LittleEndian.PutUint16(packet[3:], update.ManaCur)
	if includeMaximum {
		length = 7
		packet[0] = byte(netmsg.MSG_REPORT_TOTAL_MANA)
		binary.LittleEndian.PutUint16(packet[5:], update.ManaMax)
	}
	return sendImportantPacketWrapperC(recipient, packet[:length], nil, 1, importantPacketReplaceExisting), true
}

//export nox_xxx_netSendTeam_4D8670
func nox_xxx_netSendTeam_4D8670(recipient C.int, unit *nox_object_t) C.int {
	return C.int(GetServer().S().NetSendTeamHealth4D8670(int(recipient), asObjectS(unit)))
}

//export nox_xxx_netSendPlrHealthToTeam_4D86E0
func nox_xxx_netSendPlrHealthToTeam_4D86E0(recipient C.int) C.uintptr_t {
	return C.uintptr_t(GetServer().S().ReportPlayerHealthToTeam4D86E0(int(recipient)))
}

func ReportPlayerVitalsThrottled528030(s *server.Server, unit *server.Object) int16 {
	return s.ReportPlayerVitalsThrottled528030(unit, func(recipient byte) {
		s.ReportPlayerHealthToTeam4D86E0(int(recipient))
	}, func(recipient byte, unit *server.Object) {
		NetReportManaNative4D8930(s, recipient, unit)
	})
}

//export sub_528030
func sub_528030(unit *nox_object_t) C.short {
	return C.short(ReportPlayerVitalsThrottled528030(GetServer().S(), asObjectS(unit)))
}

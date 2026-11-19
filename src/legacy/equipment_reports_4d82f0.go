package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

const equipmentWeaponClasses = object.ClassWeapon | object.ClassWand | object.ClassFlag

func sendEquipmentReport(recipient C.int, data []byte, removeIfDisconnected C.int) C.int {
	payload := C.nox_important_payload_t(unsafe.Pointer(unsafe.SliceData(data)))
	return nox_xxx_netSendPacket1_4E5390(recipient, payload, C.int(len(data)), nil, removeIfDisconnected)
}

//export nox_xxx_netReportCharges_4D82B0
func nox_xxx_netReportCharges_4D82B0(recipient C.int, item *nox_object_t, current, maximum C.char) C.int {
	payload := [5]byte{byte(netmsg.MSG_REPORT_CHARGES), 0, 0, byte(current), byte(maximum)}
	binary.LittleEndian.PutUint16(payload[1:], uint16(server.ObjectNetCode(asObjectS(item))))
	return sendEquipmentReport(recipient, payload[:], 0)
}

//export sub_4D82F0
func sub_4D82F0(recipient C.int, citem *nox_object_t) C.intptr_t {
	item := asObjectS(citem)
	class := item.Class()
	weapon := class.HasAny(equipmentWeaponClasses)
	if !weapon && !class.Has(object.ClassArmor) {
		return C.intptr_t(uintptr(unsafe.Pointer(citem)))
	}
	attrs := item.InitDataModifier()
	modified := attrs != nil && attrs.HasModifiers()
	holder := item.InvHolder
	code := uint16(holder.NetCode)
	if holder.Class().Has(object.ClassPlayer) {
		code |= 0x8000
	}

	var payload [11]byte
	binary.LittleEndian.PutUint16(payload[1:], code)
	var flags uint32
	if weapon {
		payload[0] = byte(netmsg.MSG_REPORT_MUNDANE_WEAPON_EQUIP)
		if modified {
			payload[0] = byte(netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP)
		}
		flags = uint32(nox_xxx_weaponInventoryEquipFlags_415820(citem))
	} else {
		payload[0] = byte(netmsg.MSG_REPORT_MUNDANE_ARMOR_EQUIP)
		if modified {
			payload[0] = byte(netmsg.MSG_REPORT_MODIFIABLE_ARMOR_EQUIP)
		}
		flags = uint32(nox_xxx_unitArmorInventoryEquipFlags_415C70(citem))
	}
	binary.LittleEndian.PutUint32(payload[3:], flags)
	if !modified {
		return C.intptr_t(sendEquipmentReport(recipient, payload[:7], 0))
	}
	for i, modifier := range attrs.Modifiers {
		payload[7+i] = 0xff
		if modifier != nil {
			payload[7+i] = byte(modifier.Index())
		}
	}
	return C.intptr_t(sendEquipmentReport(recipient, payload[:], 0))
}

//export nox_xxx_netReportDequip_4D84C0
func nox_xxx_netReportDequip_4D84C0(recipient C.int, objectArg C.nox_equipment_report_object_t) C.intptr_t {
	citem := (*nox_object_t)(objectArg)
	item := asObjectS(citem)
	class := item.Class()
	weapon := class.HasAny(equipmentWeaponClasses)
	if !weapon && !class.Has(object.ClassArmor) {
		return C.intptr_t(uintptr(unsafe.Pointer(citem)))
	}

	var payload [7]byte
	binary.LittleEndian.PutUint16(payload[1:], uint16(item.InvHolder.NetCode))
	var flags uint32
	if weapon {
		payload[0] = byte(netmsg.MSG_REPORT_WEAPON_DEQUIP)
		flags = uint32(nox_xxx_weaponInventoryEquipFlags_415820(citem))
	} else {
		payload[0] = byte(netmsg.MSG_REPORT_ARMOR_DEQUIP)
		flags = uint32(nox_xxx_unitArmorInventoryEquipFlags_415C70(citem))
	}
	binary.LittleEndian.PutUint32(payload[3:], flags)
	return C.intptr_t(sendEquipmentReport(recipient, payload[:], 0))
}

//export nox_xxx_netReportEquip_4D8540
func nox_xxx_netReportEquip_4D8540(recipient C.int, item *nox_object_t, notifyAll C.int) C.intptr_t {
	payload := [3]byte{byte(netmsg.MSG_REPORT_EQUIP)}
	binary.LittleEndian.PutUint16(payload[1:], uint16(server.ObjectNetCode(asObjectS(item))))
	sendEquipmentReport(recipient, payload[:], 0)
	if notifyAll != 0 {
		return sub_4D82F0(255, item)
	}
	return 0
}

//export nox_xxx_netReportDequip_4D8590
func nox_xxx_netReportDequip_4D8590(recipient C.int, item C.nox_equipment_report_object_t) C.int {
	payload := [3]byte{byte(netmsg.MSG_REPORT_DEQUIP)}
	binary.LittleEndian.PutUint16(payload[1:], uint16(server.ObjectNetCode(asObjectS((*nox_object_t)(item)))))
	return sendEquipmentReport(recipient, payload[:], 0)
}

//export nox_xxx_netSendReportNPC_4D93A0
func nox_xxx_netSendReportNPC_4D93A0(recipient C.int, npc *nox_object_t) C.intptr_t {
	unit := asObjectS(npc)
	payload := [21]byte{byte(netmsg.MSG_REPORT_NPC)}
	code := uint16(unit.NetCode)
	if unit.Poison540 != 0 {
		code |= 0x8000
	}
	binary.LittleEndian.PutUint16(payload[1:], code)
	update := (*server.MonsterUpdateData)(unit.UpdateData)
	for i, color := range update.Color {
		pos := 3 + i*3
		payload[pos], payload[pos+1], payload[pos+2] = color.R, color.G, color.B
	}
	result := C.intptr_t(sendEquipmentReport(recipient, payload[:], 1))
	for item := unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) {
			result = sub_4D82F0(recipient, asObjectC(item))
		}
	}
	return result
}

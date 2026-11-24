package legacy

/*
#include "GAME3_2.h"
#include "cgo_const.h"

typedef const nox_object_t object_field_const_object_t;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netCreatureCmd_4D7EE0
func nox_xxx_netCreatureCmd_4D7EE0(player C.int, orderType C.char) C.int {
	return sendImportantPacketBytes(player, []byte{0xed, byte(orderType)}, true, true)
}

//export nox_xxx_netNotifyRate_4D7F10
func nox_xxx_netNotifyRate_4D7F10(player C.int) C.int {
	return sendImportantPacketBytes(player, []byte{0xec, byte(nox_xxx_rateGet_40A6C0())}, true, true)
}

func inventoryObjectPacket(op byte, item *nox_object_t) [5]byte {
	packet := [5]byte{op}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(nox_xxx_netGetUnitCodeServ_578AC0(item)))
	binary.LittleEndian.PutUint16(packet[3:5], asObjectS(item).TypeInd)
	return packet
}

//export nox_xxx_netReportPickup_4D8A60
func nox_xxx_netReportPickup_4D8A60(recipient C.int, item *nox_object_t) C.intptr_t {
	if asObjectS(item).Class().HasAny(object.Class(0x13001000)) {
		return nox_xxx_netReportModifiablePickup_4D8AD0(recipient, item)
	}
	packet := inventoryObjectPacket(75, item)
	sendImportantPacketBytes(recipient, packet[:], false, true)
	return nox_xxx_itemReportHealth_4D87A0(recipient, item)
}

//export nox_xxx_netReportModifiablePickup_4D8AD0
func nox_xxx_netReportModifiablePickup_4D8AD0(recipient C.int, item *nox_object_t) C.intptr_t {
	var packet [9]byte
	prefix := inventoryObjectPacket(76, item)
	copy(packet[:], prefix[:])
	attrs := (*server.ModifierInitData)(asObjectS(item).InitData)
	for i := 0; i < 4; i++ {
		packet[5+i] = 255
		if attrs != nil && attrs.Modifiers[i] != nil {
			packet[5+i] = byte(nox_modifier_getIndex(unsafe.Pointer(attrs.Modifiers[i])))
		}
	}
	sendImportantPacketBytes(recipient, packet[:], false, true)
	return nox_xxx_itemReportHealth_4D87A0(recipient, item)
}

func sendInventoryDrop(recipient C.int, item *nox_object_t) C.int {
	packet := inventoryObjectPacket(77, item)
	return sendImportantPacketBytes(recipient, packet[:], false, true)
}

//export nox_xxx_netReportDrop_4D8B50
func nox_xxx_netReportDrop_4D8B50(recipient C.int, object *C.object_field_const_object_t) C.int {
	return sendInventoryDrop(recipient, (*nox_object_t)(unsafe.Pointer(object)))
}

//export nox_xxx_netReportLesson_4D8EF0
func nox_xxx_netReportLesson_4D8EF0(unit *nox_object_t) C.int {
	object := asObjectS(unit)
	player := object.UpdateDataPlayer().Player
	packet := [11]byte{78}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(object.NetCode))
	binary.LittleEndian.PutUint32(packet[3:7], uint32(player.Lessons))
	binary.LittleEndian.PutUint32(packet[7:11], player.Field2140)
	return sendImportantPacketBytes(255, packet[:], true, true)
}

//export nox_xxx_netTimerStatus_4D8F50
func nox_xxx_netTimerStatus_4D8F50(recipient, status C.int) C.int {
	packet := [13]byte{0xd3}
	binary.LittleEndian.PutUint32(packet[1:5], uint32(status))
	binary.LittleEndian.PutUint32(packet[5:9], uint32(sub_40A230()))
	binary.LittleEndian.PutUint32(packet[9:13], gameFrame())
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportEnchant_4D8F90
func nox_xxx_netReportEnchant_4D8F90(recipient C.int, ref *C.uint32_t) C.int {
	unit := (*nox_object_t)(unsafe.Pointer(ref))
	packet := [7]byte{90}
	binary.LittleEndian.PutUint16(packet[1:3], uint16(nox_xxx_netGetUnitCodeServ_578AC0(unit)))
	binary.LittleEndian.PutUint32(packet[3:7], asObjectS(unit).Buffs)
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportObjHidden_4D8FD0
func nox_xxx_netReportObjHidden_4D8FD0(recipient C.int, ref *C.uint32_t) {
	unit := (*nox_object_t)(unsafe.Pointer(ref))
	obj := asObjectS(unit)
	if !obj.Class().HasAny(object.Class(0x800000)) {
		return
	}
	code := uint16(nox_xxx_netGetUnitCodeServ_578AC0(unit))
	op := byte(56)
	if obj.Flags().HasAny(object.Flags(0x1000000)) {
		op = 55
	}
	packet := [3]byte{op, byte(code), byte(code >> 8)}
	sendImportantPacketBytes(recipient, packet[:], true, false)
}

func sendJournalPacket(recipient C.int, action byte, name unsafe.Pointer, value uint16) C.int {
	packet := [68]byte{0xd5, action}
	for i := range 63 {
		next := *(*byte)(unsafe.Add(name, i))
		if next == 0 {
			break
		}
		packet[2+i] = next
	}
	binary.LittleEndian.PutUint16(packet[66:], value)
	return sendImportantPacketBytes(recipient, packet[:], true, false)
}

//export nox_xxx_netSendJournalAdd_4D9440
func nox_xxx_netSendJournalAdd_4D9440(recipient C.int, entry *C.nox_playerInfo_journal) C.int {
	return sendJournalPacket(recipient, 1, unsafe.Pointer(&entry.entry[0]), uint16(entry.field_3))
}

//export nox_xxx_netSendJournalRemove_4D94A0
func nox_xxx_netSendJournalRemove_4D94A0(recipient C.int, name *C.nox_cgo_const_char) C.int {
	return sendJournalPacket(recipient, 2, unsafe.Pointer(name), 0)
}

//export nox_xxx_netSendJournalUpdate_4D9500
func nox_xxx_netSendJournalUpdate_4D9500(recipient C.int, entry *C.nox_playerInfo_journal) C.int {
	return sendJournalPacket(recipient, 3, unsafe.Pointer(&entry.entry[0]), uint16(entry.field_3))
}

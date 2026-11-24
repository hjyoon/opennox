package legacy

/*
#include "GAME3_2.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/object"
)

//export nox_xxx_netSendFlagStatus_4D95A0
func nox_xxx_netSendFlagStatus_4D95A0(recipient C.int, id, state, flagIndex C.char, carrier C.short) C.int {
	packet := [6]byte{0xd8, byte(state), byte(id), byte(flagIndex)}
	binary.LittleEndian.PutUint16(packet[4:], uint16(carrier))
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netSendBallStatus_4D95F0
func nox_xxx_netSendBallStatus_4D95F0(recipient C.int, state C.char, code C.short) C.int {
	packet := [4]byte{0xd9, byte(state)}
	binary.LittleEndian.PutUint16(packet[2:], uint16(code))
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportSpellStat_4D9630
func nox_xxx_netReportSpellStat_4D9630(recipient, spell C.int, status C.char) C.int {
	packet := [6]byte{0: 0xdf, 5: byte(status)}
	binary.LittleEndian.PutUint32(packet[1:5], uint32(spell))
	return sendImportantPacketBytes(recipient, packet[:], true, false)
}

//export nox_xxx_netSendSecondaryWeapon_4D9670
func nox_xxx_netSendSecondaryWeapon_4D9670(recipient C.int, unit *C.uint32_t, status C.char) C.int {
	code := uint16(nox_xxx_netGetUnitCodeServ_578AC0((*nox_object_t)(unsafe.Pointer(unit))))
	packet := [4]byte{0xe0, byte(code), byte(code >> 8), byte(status)}
	return sendImportantPacketBytes(recipient, packet[:], false, true)
}

//export nox_xxx_netMsgLastQuiver_4D96B0
func nox_xxx_netMsgLastQuiver_4D96B0(recipient C.int, unit *C.uint32_t) C.int {
	code := uint16(nox_xxx_netGetUnitCodeServ_578AC0((*nox_object_t)(unsafe.Pointer(unit))))
	packet := [3]byte{0xe1, byte(code), byte(code >> 8)}
	return sendImportantPacketBytes(recipient, packet[:], false, true)
}

//export nox_xxx_netMsgInventoryLoaded_4D96E0
func nox_xxx_netMsgInventoryLoaded_4D96E0(recipient C.int) C.int {
	return sendImportantPacketBytes(recipient, []byte{113}, false, true)
}

//export nox_xxx_netFriendAddRemove_4D97A0
func nox_xxx_netFriendAddRemove_4D97A0(recipient C.int, unit *nox_object_t, add C.int) C.int {
	code := uint16(nox_xxx_netGetUnitCodeServ_578AC0(unit))
	op := byte(53)
	if add == 1 {
		op = 52
	}
	packet := [3]byte{op, byte(code), byte(code >> 8)}
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export sub_4D97E0
func sub_4D97E0(recipient C.int) C.int {
	return sendImportantPacketBytes(recipient, []byte{54}, true, true)
}

//export nox_xxx_netMsgFadeBeginPlayer
func nox_xxx_netMsgFadeBeginPlayer(recipient, direction, flag C.int) C.int {
	packet := [3]byte{0xe4, byte(bool2int(direction != 0)), byte(bool2int(flag != 0))}
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export nox_xxx_netReportDestroyObject_5289D0
func nox_xxx_netReportDestroyObject_5289D0(unit *nox_object_t) {
	obj := asObjectS(unit)
	players := &GetServer().S().Players
	for player := players.First(); player != nil; player = players.Next(player) {
		index := C.int(player.PlayerInd)
		if obj.Field37&(uint32(1)<<player.PlayerInd) != 0 {
			code := uint16(nox_xxx_netGetUnitCodeServ_578AC0(unit))
			packet := [3]byte{0x31 | byte(obj.Field5)>>6, byte(code), byte(code >> 8)}
			sendImportantPacketBytes(index, packet[:], true, false)
		}
		if obj.Class().HasAny(object.ClassMonster | object.ClassPlayer) {
			nox_xxx_netFriendAddRemove_4D97A0(index, unit, 0)
		}
	}
}

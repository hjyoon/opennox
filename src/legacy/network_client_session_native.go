package legacy

/*
#include "client__gui__guicon.h"
#include "GAME1.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func clientSessionPacketSize(op netmsg.Op) int {
	switch op {
	case netmsg.MSG_RESET_FRIENDS:
		return 1
	case netmsg.MSG_REPORT_ITEM_ENCHANTMENT, netmsg.MSG_SYSOP_RESULT:
		return 2
	case netmsg.MSG_OBJECT_FRIEND_ADD, netmsg.MSG_OBJECT_FRIEND_REMOVE:
		return 3
	case netmsg.MSG_REPORT_BALL_STATUS:
		return 4
	case netmsg.MSG_FX_WHITE_FLASH, netmsg.MSG_FORGET_DRAWABLES:
		return 5
	case netmsg.MSG_REPORT_FLAG_STATUS:
		return 6
	case netmsg.MSG_PLAYER_RESPAWN:
		return 9
	default:
		return 0
	}
}

func clientItemEnchantPacket(value byte) {
	hadLight := sub_467430()&8 != 0
	sub_467420(C.char(value))
	if !hadLight || sub_467430()&8 != 0 {
		return
	}
	player := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if player != nil && !bool(nox_client_drawable_testBuff_4356C0(player, 15)) {
		thing := nox_get_thing(int32(player.field_27))
		if thing != nil {
			nox_drawable_change_light_intensity_native(player, C.float(thing.light_intensity))
		}
	}
}

func clientSysopResultPacket(success bool) {
	key, line := "invalidPass", int32(6356)
	if success {
		key, line = "sysopAccess", 6351
	}
	var text [1024]wchar2_t
	formatWideStrings(text[:], clientDecodeString(key, line), nil, nil)
	nox_gui_console_Print_450B90(C.uchar(C.NOX_CONSOLE_RED), &text[0])
}

func clientSessionPacket(op netmsg.Op, data []byte) int {
	size := clientSessionPacketSize(op)
	if size == 0 {
		return 0
	}
	if len(data) < size {
		return -1
	}
	if op == netmsg.MSG_FORGET_DRAWABLES {
		if noxflags.HasGame(noxflags.GameOnline) {
			sub_45A670(binary.LittleEndian.Uint32(data[1:5]))
		}
		return size
	}
	if nox_client_isConnected_43C700() == 0 {
		return size
	}
	switch op {
	case netmsg.MSG_OBJECT_FRIEND_ADD:
		nox_xxx_cliAddObjFriend_4959F0(C.int(binary.LittleEndian.Uint16(data[1:3]) & 0x7fff))
	case netmsg.MSG_OBJECT_FRIEND_REMOVE:
		sub_495A20(C.int(binary.LittleEndian.Uint16(data[1:3]) & 0x7fff))
	case netmsg.MSG_RESET_FRIENDS:
		sub_4959B0()
	case netmsg.MSG_REPORT_ITEM_ENCHANTMENT:
		clientItemEnchantPacket(data[1])
	case netmsg.MSG_FX_WHITE_FLASH:
		nox_xxx____setargv_11_473920()
	case netmsg.MSG_SYSOP_RESULT:
		clientSysopResultPacket(data[1] == 1)
	case netmsg.MSG_REPORT_FLAG_STATUS:
		sub_455D80(C.uchar(data[2]), C.char(data[1]))
		sub_4705F0(C.char(data[1]), C.char(data[3]), C.short(binary.LittleEndian.Uint16(data[4:6])))
	case netmsg.MSG_REPORT_BALL_STATUS:
		sub_456140(C.uchar(data[1]))
		sub_470650(C.char(data[1]), C.short(binary.LittleEndian.Uint16(data[2:4])))
	case netmsg.MSG_PLAYER_RESPAWN:
		code := binary.LittleEndian.Uint16(data[1:3])
		if noxflags.HasGame(noxflags.GameOnline) && int(code) == ClientPlayerNetCode() {
			sub_45A670(binary.LittleEndian.Uint32(data[3:7]))
		}
		if data[8]&1 != 0 {
			nox_xxx_cliPlayerRespawn_417680(nox_common_playerInfoGetByID_417040(int32(code)), C.uint8_t(data[7]))
		}
	}
	return size
}

func clientLegacyEquipmentReport(op netmsg.Op, data []byte) int {
	size := 7
	if op == netmsg.MSG_REPORT_MODIFIABLE_WEAPON_EQUIP || op == netmsg.MSG_REPORT_MODIFIABLE_ARMOR_EQUIP {
		size = 11
	}
	if len(data) < size {
		return -1
	}
	if nox_client_isConnected_43C700() == 0 {
		return size
	}
	rawCode := binary.LittleEndian.Uint16(data[1:3])
	item := C.uint32_t(binary.LittleEndian.Uint32(data[3:7]))
	if op == netmsg.MSG_REPORT_ARMOR_DEQUIP || op == netmsg.MSG_REPORT_WEAPON_DEQUIP {
		sub_417B80(C.uint8_t(data[0]), C.int(rawCode), item)
		return size
	}
	code := rawCode & 0x7fff
	// Keep the legacy ABI's in-place netcode normalization.
	binary.LittleEndian.PutUint16(data[1:3], code)
	ids := [4]byte{0xff, 0xff, 0xff, 0xff}
	if size == 11 {
		copy(ids[:], data[7:11])
	}
	modifiers := C.nox_equipment_modifier_ids(unsafe.Pointer(&ids[0]))
	if rawCode&0x8000 != 0 {
		nox_xxx_clientEquipWeaponArmor_417AA0(C.uint8_t(data[0]), C.int(code), item, modifiers)
	} else {
		nox_xxx_clientEquip_49A3D0(C.uint8_t(data[0]), C.int(code), item, modifiers)
	}
	return size
}

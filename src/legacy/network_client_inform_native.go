package legacy

/*
#include "client__network__inform.h"
#include <string.h>

static void nox_format_inform_number(wchar2_t* out, const wchar2_t* format, int value) {
	nox_swprintf(out, format, value);
}
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

const informStringsFile = `C:\NoxPost\src\client\Network\inform.c`

var informFixedSize = [...]uint8{
	6, 6, 6, 6, 6, 10, 10, 10, 6, 10, 10, 10, 6, 6, 11, 0, 6, 2, 6, 6, 6, 6,
}

type informText struct {
	key   string
	line  int32
	sound int32
}

var informPlayerText = [...]informText{
	3:    {"netserv.c:PlayerTimeout", 147, 0},
	4:    {"objcoll.c:FlagRetrieveNotice", 164, 305},
	5:    {"objcoll.c:FlagCaptureNotice", 196, 306},
	6:    {"objcoll.c:FlagPickupNotice", 213, 303},
	7:    {"drop.c:FlagDropNotice", 230, 304},
	0x12: {"objcoll.c:PlayerExited", 247, 0},
	0x13: {"objcoll.c:PlayerExitedWarp", 265, 0},
	0x14: {"GeneralPrint:SecretFoundOther", 90, 0},
}

var informEffectText = [...]informText{
	{"atckexec.c:PlayerStunned", 27, 0},
	{"atckexec.c:PlayerConfused", 31, 0},
	{"atckexec.c:PlayerPoisoned", 35, 0},
	{"player.c:TooHeavy", 39, 0},
}

func informString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(informStringsFile), line)
}

func informPrintTemplate(format, name, team *wchar2_t) {
	var text [256]wchar2_t
	formatWideStrings(text[:], format, name, team)
	nox_xxx_printCentered_445490(&text[0])
}

func informPrint(key string, line int32, name, team *wchar2_t) {
	informPrintTemplate(informString(key, line), name, team)
}

func clientInformPacket(data []byte) int {
	if len(data) < 2 || int(data[1]) >= len(informFixedSize) {
		return 0
	}
	subtype := data[1]
	size := int(informFixedSize[subtype])
	if subtype == 0xF {
		if len(data) < 4 {
			return 0
		}
		end := bytes.IndexByte(data[3:], 0)
		if end < 0 {
			return 0
		}
		size = 4 + end
		if nox_xxx_gameGetPlayState_4356B0() == 3 {
			key := string(data[3 : size-1])
			if data[2] != 0 {
				value := informString(key, 397)
				informPrint("use.c:SignSays", 396, value, nil)
			} else {
				nox_xxx_printCentered_445490(informString(key, 402))
			}
		}
		return size
	}
	if len(data) < size {
		return 0
	}
	var value uint32
	if size >= 6 {
		value = binary.LittleEndian.Uint32(data[2:6])
	}
	switch subtype {
	case 0:
		nox_xxx_abilGetError_4FB0B0_magic_plyrspel(C.uint32_t(value))
	case 1:
		title := nox_xxx_spellTitle_424930(int32(value))
		informPrint("plyrspel.c:SpellCastSuccess", 117, title, nil)
	case 2:
		nox_xxx_abilGetSuccess_4FB960_ability(C.uint32_t(value))
	case 3, 4, 5, 6, 7, 0x12, 0x13, 0x14:
		player := nox_common_playerInfoGetByID_417040(int32(value))
		if player == nil {
			return size
		}
		var team *wchar2_t
		if size == 10 {
			team = nox_server_teamTitle_418C20(int32(binary.LittleEndian.Uint32(data[6:10])))
		}
		message := informPlayerText[subtype]
		informPrint(message.key, message.line, &player.name_final[0], team)
		if message.sound != 0 {
			nox_xxx_clientPlaySoundSpecial_452D80(C.int(message.sound), 100)
		}
	case 8:
		title := nox_server_teamTitle_418C20(int32(value))
		informPrint("update.c:FlagRespawnNotice", 179, title, nil)
		nox_xxx_clientPlaySoundSpecial_452D80(305, 100)
	case 9:
		team := nox_xxx_getTeamByID_418AB0(int32(binary.LittleEndian.Uint32(data[6:10])))
		player := nox_common_playerInfoGetByID_417040(int32(value))
		if team != nil {
			if player == nil {
				informPrint("objcoll.c:FlagBallUnknownNotice", 374, &team.name[0], nil)
			} else {
				informPrint("objcoll.c:FlagBallNotice", 365, &player.name_final[0], &team.name[0])
			}
		}
	case 0xA, 0xB:
		player := nox_common_playerInfoGetByID_417040(int32(value))
		team := nox_xxx_getTeamByID_418AB0(int32(binary.LittleEndian.Uint32(data[6:10])))
		if player == nil {
			return size
		}
		if subtype == 0xA {
			if team == nil {
				informPrint("pickup.c:PickUpCrown", 343, &player.name_final[0], nil)
			} else {
				informPrint("pickup.c:PickUpTeamCrown", 334, &player.name_final[0], &team.name[0])
			}
		} else if team == nil {
			informPrint("drop.c:DropCrown", 312, &player.name_final[0], nil)
		} else {
			informPrint("drop.c:DropTeamCrown", 303, &player.name_final[0], &team.name[0])
		}
	case 0xC:
		nox_xxx_printCentered_445490(informString("Netserv.c:InObservationMode", 128))
		if value != 0 {
			nox_xxx_printCentered_445490(informString("Netserv.c:PressJump", 132))
		}
	case 0xD:
		if value < uint32(len(informEffectText)) {
			message := informEffectText[value]
			nox_xxx_printCentered_445490(informString(message.key, message.line))
		}
	case 0xE:
		sub_495210((*C.uint8_t)(unsafe.Pointer(&data[0])))
	case 0x10:
		title := nox_server_teamTitle_418C20(int32(value))
		informPrint("pickup.c:WrongTeam", 105, title, nil)
	case 0x11:
		message := informString("Noxworld.c:ErrChangedClass", 51)
		title := informString("guiserv.c:Notice", 50)
		nox_xxx_dialogMsgBoxCreate_449A10(nil, title, message, 33, nil, nil)
	case 0x15:
		var text [256]wchar2_t
		C.nox_format_inform_number(&text[0], informString("GeneralPrint:AdvanceToStage1", 281), C.int(value))
		informPrint("use.c:SignSays", 282, &text[0], nil)
	}
	return size
}

//export nox_client_handlePacketInformSized_4C9BF0
func nox_client_handlePacketInformSized_4C9BF0(data *C.uint8_t, size C.int) C.int {
	if data == nil || size < 0 {
		return 0
	}
	return C.int(clientInformPacket(unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))))
}

//export nox_client_handlePacketInform_4C9BF0
func nox_client_handlePacketInform_4C9BF0(data *C.uint8_t) C.int {
	if data == nil {
		return 0
	}
	subtype := *(*byte)(unsafe.Add(unsafe.Pointer(data), 1))
	if int(subtype) >= len(informFixedSize) {
		return 0
	}
	size := C.int(informFixedSize[subtype])
	if subtype == 0xF {
		key := (*C.char)(unsafe.Add(unsafe.Pointer(data), 3))
		size = C.int(C.strlen(key)) + 4
	}
	return nox_client_handlePacketInformSized_4C9BF0(data, size)
}

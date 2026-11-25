package legacy

/*
#include "client__network__cdecode.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_1200768, dword_5d4594_1200832;
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const clientSettingsRecordSize = 58
const clientDecodeStringsFile = `C:\NoxPost\src\Client\Network\cdecode.c`

func clientSettingsRecord(index byte) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(nox_xxx_cliGamedataGet_416590(C.int(index)))), clientSettingsRecordSize)
}

func clientDecodeString(key string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(clientDecodeStringsFile), line)
}

func clientGameSettingsPacket(data []byte) {
	C.dword_5d4594_1200768 = 0
	frame := binary.LittleEndian.Uint32(data[1:5])
	if uint32(nox_xxx_get3512_40A350()) >= frame {
		return
	}
	nox_xxx_set3512_40A340(C.int(frame))
	hadCTF := noxflags.HasGame(noxflags.GameModeCTF)
	hadQuest := noxflags.HasGame(noxflags.GameModeQuest)
	if noxflags.HasGame(noxflags.GameHost) {
		return
	}
	noxflags.UnsetGame(noxflags.GameFlag(524272))
	noxflags.SetGame(noxflags.GameFlag(binary.LittleEndian.Uint32(data[9:13])))
	sub_409E40(C.int(binary.LittleEndian.Uint32(data[13:17])))
	nox_client_setVersion_409AE0(binary.LittleEndian.Uint32(data[5:9]))
	nox_xxx_servSetPlrLimit_409F80(C.int(data[17]))

	settings := clientSettingsRecord(0)
	if int(int8(settings[56])) != int(data[19]) || binary.LittleEndian.Uint16(settings[54:56]) != uint16(data[18]) {
		C.dword_5d4594_1200768 = 1
	}
	binary.LittleEndian.PutUint16(settings[54:56], uint16(data[18]))
	settings[56] = data[19]
	if int8(data[9]) >= 0 {
		copy(settings[52:54], data[9:11])
		name := GoString(nox_xxx_mapGetMapName_409B40())
		end := copy(settings, name)
		if end < len(settings) {
			settings[end] = 0
		}
	}
	if !noxflags.HasGame(noxflags.GameModeCTF) {
		sub_455C10()
	} else if !hadCTF {
		sub_455A50(2)
	}
	// The original uses the previous CTF bit for both entry transitions.
	if !noxflags.HasGame(noxflags.GameModeFlagBall) {
		sub_456050()
	} else if !hadCTF {
		sub_455F60()
	}
	if hadQuest && !noxflags.HasGame(noxflags.GameModeQuest) {
		for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
			nox_xxx_cliPlayerRespawn_417680(player, 255)
		}
		if !noxflags.HasGame(noxflags.GameFlag18) && nox_xxx_gameGetPlayState_4356B0() == 3 {
			nox_game_exit_xxx2()
		}
	}
}

func clientGameSettings2Packet(data []byte) {
	if noxflags.HasGame(noxflags.GameHost) {
		return
	}
	settings := clientSettingsRecord(0)
	name := data[1:16]
	if end := bytes.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}
	clear(settings[9:24])
	copy(settings[9:24], name)

	snapshot := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 1200708)), clientSettingsRecordSize)
	if !bytes.Equal(snapshot[24:52], data[17:45]) {
		C.dword_5d4594_1200768 = 1
	}
	copy(settings[24:44], data[17:37])
	sub_4540E0((*C.nox_cgo_const_void)(unsafe.Pointer(&settings[24])))
	copy(settings[44:52], data[37:45])
	if limit := binary.LittleEndian.Uint32(data[45:49]); limit != 0 {
		sub_40A1F0(1)
		sub_40A310(C.int(limit))
	} else {
		sub_40A1F0(0)
	}
	copy(snapshot, settings)
	if nox_client_isConnected_43C700() != 0 && C.dword_5d4594_1200768 != 0 {
		nox_xxx_printCentered_445490(clientDecodeString("OptionsChanged", 1429))
		nox_xxx_clientPlaySoundSpecial_452D80(310, 100)
	}
}

func clientGUISettingsPacket(data []byte) {
	index := data[1]
	settings := clientSettingsRecord(index)
	oldFlags := settings[52]
	copy(settings, data[2:60])
	if (oldFlags^data[54])&(1<<5) != 0 {
		sub_4573A0()
	}
	if sub_4169C0() == 0 {
		if sub_459DA0() == 0 {
			if index == 1 {
				sub_4165F0(1, 0)
			}
			if noxflags.HasGame(noxflags.GameModeChat) {
				if C.dword_5d4594_1200832 != 0 {
					player := (*C.nox_playerInfo)(Get_dword_8531A0_2576().C())
					var text [128]wchar2_t
					formatWideStrings(text[:], clientDecodeString("NameChange", 1470), &player.name_final[0], nil)
					nox_xxx_printCentered_445490(&text[0])
					C.dword_5d4594_1200832 = 0
				}
				if Get_nox_server_sanctuaryHelp_54276() != 0 {
					nox_xxx_cliShowHelpGui_49C560()
				}
			}
		}
		nox_xxx_cliSetSettingsAcquired_4169D0(1)
	}
	sub_459C30()
}

func clientSettingsPacket(op netmsg.Op, data []byte) (int, bool) {
	switch op {
	case netmsg.MSG_GAME_SETTINGS:
		if len(data) >= 20 {
			clientGameSettingsPacket(data)
			return 20, true
		}
	case netmsg.MSG_GAME_SETTINGS_2:
		if len(data) >= 49 {
			clientGameSettings2Packet(data)
			return 49, true
		}
	case netmsg.MSG_UPDATE_GUI_GAME_SETTINGS:
		if len(data) >= 60 {
			clientGUISettingsPacket(data)
			return 60, true
		}
	default:
		return 0, false
	}
	return -1, true
}

//export nox_client_decodeSettingsPacket
func nox_client_decodeSettingsPacket(op C.int, data *C.uchar, size C.int) C.int {
	if data == nil || size < 0 {
		return -1
	}
	n, _ := clientSettingsPacket(netmsg.Op(op), unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size)))
	return C.int(n)
}

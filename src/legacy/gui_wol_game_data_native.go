package legacy

/*
#include "GAME1_2.h"

extern uint32_t dword_5d4594_815052;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_43AA70
func sub_43AA70() *C.char {
	reconnecting := Get_dword_5d4594_528256() != 0
	if Get_dword_5d4594_528252() != 0 && reconnecting {
		NetworkLogPrint("RECON: Posting server to WOL")
	}
	Set_nox_game_createOrJoin_815048(0)
	C.dword_5d4594_815052 = 1

	current := nox_xxx_cliGamedataGet_416590(0)
	record := unsafe.Slice((*byte)(sub_416640()), int(C.sizeof_nox_gui_server_ent_t))
	copy(record[111:], unsafe.Slice((*byte)(unsafe.Pointer(current)), len(record)-111))
	binary.LittleEndian.PutUint16(record[163:], uint16(noxflags.GetGame()))
	for off := 135; off < 163; off++ {
		record[off] = 0xff
	}

	serverName := record[120:135]
	clear(serverName)
	copy(serverName, GoString(nox_xxx_serverOptionsGetServername_40A4C0()))
	mapName := record[111:120]
	length := copy(mapName[:len(mapName)-1], GoString(nox_xxx_mapGetMapName_409B40()))
	mapName[length] = 0
	if nox_xxx_isQuest_4D6F50() != 0 {
		stage := uint16(1)
		if reconnecting {
			stage = uint16(nox_game_getQuestStage_4E3CC0())
		}
		binary.LittleEndian.PutUint16(record[165:], stage)
	}
	record[104] = byte(nox_xxx_servGetPlrLimit_409FA0())
	record[103] = byte(nox_common_playerInfoCount_416F40())
	if noxflags.HasEngine(noxflags.EngineNoRendering) {
		record[103]--
		record[104]--
	}
	mode := sub_43BE50_get_video_mode_id()
	binary.LittleEndian.PutUint32(record[48:], uint32(C.NOX_CLIENT_VERS_CODE))
	record[102] = record[102]&0x80 | byte(mode)
	binary.LittleEndian.PutUint32(record[44:], memmap.Uint32(0x5D4594, 814916))
	binary.LittleEndian.PutUint16(record[109:], uint16(nox_xxx_servGetPort_40A430()))
	ClientSetServerHost("localhost")

	result := nox_xxx_cliGamedataGet_416590(1)
	flags := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(result), 52)), 2)
	binary.LittleEndian.PutUint16(flags, binary.LittleEndian.Uint16(flags)&0xe90f|0x0100)
	return result
}

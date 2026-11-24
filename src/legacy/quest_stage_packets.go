package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	questStageFirstNameOffset  = 3838
	questStageSecondNameOffset = 3806
)

func writeQuestStageName(dst []byte, offset uintptr) {
	nox_server_currentMapGetFilename_409B30()
	source := memmap.Slice(0x973F18, offset)[:len(dst)]
	alloc.StrCopyZero(dst, alloc.GoStringS(source))
}

func sendQuestStagePacket(recipient C.int, kind byte, ready bool) C.int {
	packet := [69]byte{byte(netmsg.MSG_GAUNTLET), kind}
	if ready {
		packet[4] |= 1
	}
	if sub_51A950() != 0 {
		packet[4] |= 2
	}
	binary.LittleEndian.PutUint16(packet[2:4], uint16(nox_game_getQuestStage_4E3CC0()))
	writeQuestStageName(packet[5:37], questStageFirstNameOffset)
	writeQuestStageName(packet[37:69], questStageSecondNameOffset)
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export sub_4D6880
func sub_4D6880(recipient C.int, ready C.int) C.int {
	return sendQuestStagePacket(recipient, 13, ready != 0)
}

//export sub_4D6940
func sub_4D6940() *C.char {
	return (*C.char)(memmap.PtrOff(0x973F18, questStageFirstNameOffset))
}

//export sub_4D6950
func sub_4D6950() *C.char {
	return (*C.char)(memmap.PtrOff(0x973F18, questStageSecondNameOffset))
}

//export nox_game_sendQuestStage_4D6960
func nox_game_sendQuestStage_4D6960(recipient C.int) C.int {
	return sendQuestStagePacket(recipient, 14, false)
}

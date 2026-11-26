package legacy

/*
#include "GAME3_2.h"

*/
import "C"

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/common/memmap"
)

func legacyQuestStateWord(offset uintptr) *uint32 {
	return memmap.PtrUint32(0x5D4594, offset)
}

func setLegacyQuestStateWord(offset uintptr, value C.int) C.int {
	*legacyQuestStateWord(offset) = uint32(value)
	return value
}

//export nox_xxx_isQuest_4D6F50
func nox_xxx_isQuest_4D6F50() C.int {
	return C.int(int32(*legacyQuestStateWord(1556160)))
}

//export nox_xxx_setQuest_4D6F60
func nox_xxx_setQuest_4D6F60(value C.int) C.int {
	return setLegacyQuestStateWord(1556160, value)
}

//export sub_4D6F70
func sub_4D6F70() C.int {
	return C.int(int32(*legacyQuestStateWord(1556164)))
}

//export sub_4D6F80
func sub_4D6F80(value C.int) C.int {
	return setLegacyQuestStateWord(1556164, value)
}

//export sub_4D6FA0
func sub_4D6FA0() C.int {
	return C.int(int32(*legacyQuestStateWord(1556104)))
}

//export sub_4D71E0
func sub_4D71E0(deadline C.int) C.int {
	legacyWords.dword_5d4594_1556136 = C.uint32_t(deadline)
	return deadline
}

//export sub_4D7280
func sub_4D7280(recipient C.int, status C.char) C.int {
	packet := [3]byte{byte(netmsg.MSG_GAUNTLET), 24, byte(status)}
	return sendImportantPacketBytes(recipient, packet[:], true, true)
}

//export sub_4D72D0
func sub_4D72D0(stage C.int) C.int {
	previous := uint32(legacyWords.dword_5d4594_1556128)
	*legacyQuestStateWord(1556132) = previous
	legacyWords.dword_5d4594_1556128 = C.uint32_t(stage)
	return C.int(int32(previous))
}

//export sub_4D7300
func sub_4D7300() C.int {
	return C.int(int32(*legacyQuestStateWord(1556132)))
}

//export sub_4D7430
func sub_4D7430() C.int {
	return C.int(int32(*legacyQuestStateWord(1556116)))
}

//export sub_4D7440
func sub_4D7440(value C.int) C.int {
	return setLegacyQuestStateWord(1556116, value)
}

//export sub_4D7450
func sub_4D7450(recipient C.int, stage C.short) C.int {
	word := uint16(stage)
	packet := [4]byte{byte(netmsg.MSG_GAUNTLET), 29, byte(word), byte(word >> 8)}
	return sendImportantPacketBytes(recipient, packet[:], true, false)
}

//export sub_4D75F0
func sub_4D75F0(frame C.int) C.int {
	return setLegacyQuestStateWord(1556108, frame)
}

//export sub_4D76E0
func sub_4D76E0(value C.int) C.int {
	return setLegacyQuestStateWord(1556124, value)
}

//export sub_4D76F0
func sub_4D76F0() C.int {
	return C.int(int32(*legacyQuestStateWord(1556124)))
}

//export sub_4D79A0
func sub_4D79A0(player C.char) C.int {
	// GAME.EXE uses SHL EAX, CL, which masks the count to five bits.
	mask := ^(uint32(1) << (uint8(player) & 31))
	*legacyQuestStateWord(1556300) &= mask
	return C.int(int32(mask))
}

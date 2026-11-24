package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//export sub_4D61F0
func sub_4D61F0(unit *nox_object_t) *nox_playerInfo {
	playerUnit := asObjectS(unit)
	if playerUnit == nil || playerUnit.Flags().Has(object.FlagDestroyed) {
		return nil
	}
	player := (*server.PlayerUpdateData)(playerUnit.UpdateData).Player
	player.RecordSecretFound()
	return (*nox_playerInfo)(player.C())
}

func Sub_4D70B0() {
	questClosingSettings4D70B0()
}

//export nox_quest_closing_settings_4D70B0
func nox_quest_closing_settings_4D70B0() C.uintptr_t {
	return C.uintptr_t(questClosingSettings4D70B0())
}

func questClosingSettings4D70B0() uintptr {
	if state := memmap.Uint32(0x5D4594, 1556152); state != 0 {
		return uintptr(state)
	}
	return uintptr(sub4169F0Native())
}

//export nox_broadcast_player_text_4D9FD0
func nox_broadcast_player_text_4D9FD0(tag C.char, text *wchar2_t) C.int {
	var buffer [11 + 2*255]byte
	packet := broadcastPlayerTextPacket4D9FD0(&buffer, byte(tag), (*uint16)(unsafe.Pointer(text)))
	world := GetServer().S()
	for unit := world.Players.FirstUnit(); unit != nil; unit = world.NextPlayerUnit4DA7F0(unit) {
		if update := (*server.PlayerUpdateData)(unit.UpdateData); update != nil && update.Player != nil {
			world.NetList.AddToMsgListCli(update.Player.PlayerIndex(), netlist.Kind1, packet)
		}
	}
	return 0
}

func broadcastPlayerTextPacket4D9FD0(buffer *[11 + 2*255]byte, tag byte, text *uint16) []byte {
	buffer[0] = byte(netmsg.MSG_TEXT_MESSAGE)
	if cliCanTalkNative4100F0(text) != 0 {
		tag |= 2
	} else {
		tag |= 4
	}
	buffer[3] = tag
	count := byte(alloc.StrLen(text) + 1)
	buffer[8] = count
	words := unsafe.Slice(text, int(count))
	if tag&4 != 0 {
		for i, word := range words {
			binary.LittleEndian.PutUint16(buffer[11+2*i:], word)
		}
		return buffer[:11+2*int(count)]
	}
	for i, word := range words {
		buffer[11+i] = byte(word)
	}
	return buffer[:11+int(count)]
}

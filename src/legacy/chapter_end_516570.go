package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4_1.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

var sendChapterEndPacket4D9560 = func(recipient int32, packet [3]byte) int32 {
	return int32(nox_xxx_netSendPacket0_4E5420(C.int(recipient), C.nox_important_payload_t(unsafe.Pointer(&packet[0])), 3, nil, 1))
}

//export nox_xxx_netSendChapterEnd_4D9560
func nox_xxx_netSendChapterEnd_4D9560(recipient C.int, stage C.char, endState C.int) C.int {
	packet := [3]byte{0xd6, byte(stage)}
	if endState == 1 {
		packet[2] = 1
	}
	return C.int(sendChapterEndPacket4D9560(int32(recipient), packet))
}

type chapterEndUnits516570 struct {
	first func() *server.Object
	next  func(*server.Object) *server.Object
}

var chapterEndPlayerUnits516570 = func() chapterEndUnits516570 {
	s := GetServer().S()
	return chapterEndUnits516570{first: s.Players.FirstUnit, next: s.Players.NextUnit}
}

//export sub_516570
func sub_516570() C.int {
	units := chapterEndPlayerUnits516570()
	var player *server.Player
	for unit := units.first(); unit != nil; unit = units.next(unit) {
		player = nil
		if update := unit.UpdateDataPlayer(); update != nil {
			player = update.Player
		}
		if player != nil && player.PlayerInd == server.HostPlayerIndex {
			break
		}
	}
	stage := memmap.Uint8(0x5D4594, 2386828)
	endState := memmap.Int32(0x5D4594, 2386832)
	Set_nox_gameDisableMapDraw_5d4594_2650672(1)
	if player == nil {
		return 0
	}
	return nox_xxx_netSendChapterEnd_4D9560(C.int(player.PlayerInd), C.char(stage), C.int(endState))
}

package legacy

/*
#include <stdint.h>
extern uint32_t dword_5d4594_739392;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The kill-pair cursor is a relocated C-owned scalar, not the retired blob
// slot at 5D4594+739392. No player pointer crosses these scalar accessors.
func playerKillStatsPairCount425CA0() uint32 {
	return uint32(C.dword_5d4594_739392)
}

func playerKillStatsSetPairCount425CA0(value uint32) {
	C.dword_5d4594_739392 = C.uint32_t(value)
}

func playerKillStatsRuntime425CA0() server.PlayerKillStatsRuntime425CA0 {
	return server.PlayerKillStatsRuntime425CA0{
		GameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		PlayerCount: Get_dword_5d4594_608316,
		SetCount:    Set_dword_5d4594_608316,
		CopyName: func(index uint32, player *server.Player) {
			name := &player.Field2096Buf[0]
			destination := (*byte)(memmap.PtrOff(0x5D4594, uintptr(uint32(600124)+index*32)))
			// Like the original strcpy, copy through the terminator without
			// clearing the remaining record bytes or imposing a new name cap.
			alloc.StrCopyP(unsafe.Slice(destination, alloc.StrLen(name)+1), name)
		},
		ConnectionIP: func(connection int) uint32 {
			return nox_xxx_net_getIP_554200(int32(connection))
		},
		StoreIP: func(index, value uint32) {
			*memmap.PtrUint32(0x5D4594, uintptr(uint32(600136)+index*32)) = value
		},
		StoreTeam: func(index, value uint32) {
			*memmap.PtrUint32(0x5D4594, uintptr(uint32(600140)+index*32)) = value
		},
		StoreClass: func(index uint32, value byte) {
			*memmap.PtrUint8(0x5D4594, uintptr(uint32(600144)+index*32)) = value
		},
		PairCount: playerKillStatsPairCount425CA0,
		StorePair: func(index uint32, first, second byte) {
			*memmap.PtrUint8(0x5D4594, uintptr(uint32(608320)+index*2)) = first
			*memmap.PtrUint8(0x5D4594, uintptr(uint32(608321)+index*2)) = second
		},
		SetPairCount: playerKillStatsSetPairCount425CA0,
		// This is the original external bulk-report service boundary, not
		// a claim that its remaining ABI32 internals have been ported.
		Flush: Nox_xxx_net_4263C0,
	}
}

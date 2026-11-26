package legacy

/*
#include "GAME1_1.h"
#include "stats_bulk_native.h"
#include "stats_start_native.h"
#include "stats_end_native.h"

extern uint32_t dword_5d4594_600116;
*/
import "C"

import (
	"encoding/binary"
	"math/bits"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func updateStatsPlayerByte(player *nox_playerInfo, value C.char, base uintptr) C.int {
	if !noxflags.HasGame(noxflags.GameOnline) {
		return 0
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		return 1
	}
	if player == nil {
		return 0
	}
	index := asPlayerS(player).Field4648
	if index == -1 {
		return -1
	}
	relative := 32 * index
	*memmap.PtrUint8(0x5D4594, base+uintptr(relative)) = byte(value)
	return C.int(relative)
}

//export sub_425E90
func sub_425E90(player *C.nox_playerInfo, value C.char) C.int {
	return updateStatsPlayerByte(player, value, 600145)
}

//export sub_425ED0
func sub_425ED0(player *C.nox_playerInfo, value C.char) C.int {
	return updateStatsPlayerByte(player, value, 600152)
}

//export sub_425F10
func sub_425F10(raw *C.nox_playerInfo) {
	if !noxflags.HasGame(noxflags.GameOnline) || noxflags.HasGame(noxflags.GameModeQuest) || raw == nil {
		return
	}
	player := asPlayerS(raw)
	if player.Field4648 != -1 {
		return
	}
	index := Get_dword_5d4594_608316()
	Set_dword_5d4594_608316(index + 1)
	copyStatsPlayerName(index, player)
	base := 600124 + uintptr(32*index)
	connection := int32(player.PlayerInd) + 1
	if player.PlayerInd == 31 {
		connection = 0
	}
	*memmap.PtrUint32(0x5D4594, base+12) = bits.ReverseBytes32(nox_xxx_net_getIP_554200(connection))
	dedicatedHost := player.PlayerInd == 31 && noxflags.HasEngine(noxflags.EngineNoRendering)
	observer := player.Field3680&1 != 0 && player.Field3680&0x20 == 0
	*memmap.PtrUint8(0x5D4594, base+28) = byte(bool2int(!dedicatedHost && !observer))
	*memmap.PtrUint32(0x5D4594, base+16) = player.Field2068
	*memmap.PtrUint8(0x5D4594, base+20) = byte(player.Info().PlayerClass())
	*memmap.PtrUint32(0x5D4594, base+24) = uint32(time.Now().Unix())
	*memmap.PtrUint8(0x5D4594, base+21) = 1
	player.Field4648 = int32(index)
}

//export sub_426150
func sub_426150() {
	if !noxflags.HasGame(noxflags.GameOnline) || noxflags.HasGame(noxflags.GameModeQuest) {
		return
	}
	fillStatsStartHeader((*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_599496)))
}

func fillStatsStartHeader(modeCell *uint32) {
	settings := (*[109]byte)(sub_416640())
	data := (*[124]byte)(unsafe.Pointer(nox_xxx_cliGamedataGet_416590(0)))
	*memmap.PtrUint16(0x5D4594, 599482) = uint16(settings[103])
	*memmap.PtrUint32(0x5D4594, 599484) = uint32(settings[104])
	*memmap.PtrUint32(0x5D4594, 599488) = binary.LittleEndian.Uint32(settings[40:])
	*memmap.PtrUint32(0x5D4594, 599492) = uint32(sub_4200E0())
	*memmap.PtrUint8(0x5D4594, 599502) = byte(bool2int(data[53]&0xc0 != 0))
	mode := binary.LittleEndian.Uint16(data[52:])
	for report, flag := range [...]uint16{0x100, 0x20, 0x40, 0x10, 0x400} {
		if mode&flag != 0 {
			*modeCell = uint32(report)
			break
		}
	}
	*memmap.PtrUint8(0x5D4594, 599500) = byte(bool2int(data[53]&0x40 == 0))
	*memmap.PtrUint32(0x5D4594, 599508) = uint32(binary.LittleEndian.Uint16(data[54:]))
	*memmap.PtrUint32(0x5D4594, 599512) = uint32(data[56])
	*memmap.PtrUint8(0x5D4594, 599516) = settings[100]
	*memmap.PtrUint32(0x5D4594, 599520) = uint32(settings[101] & 0xf)
	*memmap.PtrUint32(0x5D4594, 599524) = uint32(settings[101] >> 4)
	*memmap.PtrUint32(0x5D4594, 599528) = uint32(binary.LittleEndian.Uint16(settings[105:]))
	*memmap.PtrUint32(0x5D4594, 599532) = uint32(binary.LittleEndian.Uint16(settings[107:]))
	*memmap.PtrUint8(0x5D4594, 599536) = settings[102]
	*memmap.PtrUint8(0x5D4594, 599537) = settings[100] & 0x30
	*memmap.PtrUint8(0x5D4594, 599501) = byte(sub_417DE0())
	destination := memmap.PtrT[[16]byte](0x5D4594, 599828)
	clear(destination[:])
	copy(destination[:15], data[9:9+alloc.StrLenS(data[9:24])])
	copy(memmap.PtrT[[100]byte](0x5D4594, 599540)[:], data[24:])
	*memmap.PtrUint32(0x5D4594, 599564) = binary.LittleEndian.Uint32(data[48:])
	*memmap.PtrUint32(0x5D4594, 599560) = binary.LittleEndian.Uint32(data[44:])
	name := memmap.PtrT[[9]byte](0x5D4594, 599572)
	clear(name[:])
	copy(name[:8], data[:alloc.StrLenS(data[:8])])
	for _, address := range [...]uintptr{600112, 600084, 600088, 600092, 600096} {
		*memmap.PtrUint32(0x5D4594, address) = 0
	}
}

//export nox_xxx_net_4263C0
func nox_xxx_net_4263C0() *C.char {
	return nox_stats_bulk_flush_native()
}

//export sub_4264D0
func sub_4264D0() C.int {
	if noxflags.HasGame(noxflags.GameModeQuest) {
		*memmap.PtrUint32(0x5D4594, 739416) = uint32(time.Now().Unix()) - uint32(C.dword_5d4594_600116)
		return nox_stats_quest_end_native()
	}
	if noxflags.HasGame(noxflags.GameOnline) {
		for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
			sub_425E90(player, 1)
		}
		*memmap.PtrUint32(0x5D4594, 599504) = uint32(time.Now().Unix()) - uint32(C.dword_5d4594_600116)
	}
	return nox_stats_session_report_native(2)
}

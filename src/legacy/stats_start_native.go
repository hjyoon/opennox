package legacy

/*
#include <stdlib.h>
#include "GAME1.h"
#include "stats_end_native.h"

*/
import "C"

import (
	"encoding/binary"
	"time"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	questStatsHeaderBase = 739396
	questStatsHeaderSize = 580
)

var (
	questStatsColumns, _ = alloc.New(C.nox_stats_quest_columns_native{})
	questStatsNameCount  int
)

func questStatsHeader() []byte {
	return memmap.Slice(0x5D4594, questStatsHeaderBase)[:questStatsHeaderSize]
}

func statsCopyStringField(dst, src []byte) {
	n := copy(dst, src[:alloc.StrLenS(src)])
	clear(dst[n:])
}

//export nox_stats_start_header_native
func nox_stats_start_header_native() {
	if noxflags.HasGame(noxflags.GameOnline) && !noxflags.HasGame(noxflags.GameModeQuest) {
		fillStatsStartHeader(memmap.PtrUint32(0x5D4594, 599496))
	}
}

//export nox_stats_quest_columns_native_get
func nox_stats_quest_columns_native_get() *C.nox_stats_quest_columns_native {
	return questStatsColumns
}

//export nox_stats_quest_columns_clear_native
func nox_stats_quest_columns_clear_native() unsafe.Pointer {
	columns := nox_stats_quest_columns_native_get()
	if columns.names != nil {
		names := unsafe.Slice(columns.names, questStatsNameCount)
		for i, name := range names {
			C.free(unsafe.Pointer(name))
			names[i] = nil
		}
		C.free(unsafe.Pointer(columns.names))
		columns.names = nil
	}
	questStatsNameCount = 0
	C.free(unsafe.Pointer(columns.ips))
	C.free(unsafe.Pointer(columns.classes))
	columns.ips, columns.classes = nil, nil
	lastScoreColumn := unsafe.Pointer(columns.scores[7])
	for i, score := range columns.scores {
		C.free(unsafe.Pointer(score))
		columns.scores[i] = nil
	}
	return lastScoreColumn
}

//export sub_426060
func sub_426060() {
	Set_dword_5d4594_608316(0)
	legacyWords.dword_5d4594_600116 = C.uint32_t(time.Now().Unix())
	port := sub_5545A0()
	sub_4282D0(sub_554230(), C.int(port))
	if noxflags.HasGame(noxflags.GameModeQuest) {
		settings := (*[44]byte)(sub_416640())
		game := (*[54]byte)(unsafe.Pointer(nox_xxx_cliGamedataGet_416590(0)))
		header := questStatsHeader()
		binary.NativeEndian.PutUint16(header, uint16(sub_40A770()))
		binary.NativeEndian.PutUint32(header[4:], binary.NativeEndian.Uint32(settings[40:]))
		binary.NativeEndian.PutUint32(header[8:], uint32(sub_4200E0()))
		header[16] = byte(bool2int(game[53]&0xc0 != 0))
		binary.NativeEndian.PutUint32(header[12:], 5)
		statsCopyStringField(header[280:296], game[9:24])
		statsCopyStringField(header[24:33], game[:8])
		nox_stats_quest_columns_clear_native()
		return
	}
	if !noxflags.HasGame(noxflags.GameOnline) {
		return
	}
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		player.field_4648 = -1
	}
	if host := nox_common_playerInfoFromNum_417090(31); host != nil {
		sub_425F10(host)
	}
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		if player.playerInd != 31 {
			sub_425F10(player)
		}
	}
	nox_stats_start_header_native()
	nox_stats_session_report_native(0)
	*memmap.PtrUint16(0x5D4594, 599482) = 0
}

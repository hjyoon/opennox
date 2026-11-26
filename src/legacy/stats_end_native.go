package legacy

/*
#include <stdlib.h>
#include "stats_end_native.h"

*/
import "C"

import (
	"encoding/binary"
	"math/bits"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func questStatsPlayerCount(header []byte) int {
	return int(int16(binary.NativeEndian.Uint16(header)))
}

func buildQuestStatsPlayers(columns *C.nox_stats_quest_columns_native, header []byte) {
	binary.NativeEndian.PutUint16(header, uint16(nox_xxx_player_4E3CE0()))
	count := questStatsPlayerCount(header)
	if count > 0 {
		nox_stats_quest_columns_clear_native()
		questStatsNameCount = count
		columns.names = (**C.char)(C.calloc(C.size_t(count), C.size_t(unsafe.Sizeof((*C.char)(nil)))))
		names := unsafe.Slice(columns.names, count)
		// The stored player name can occupy the full fixed-width field.
		const nameSize = len(nox_playerInfo{}.field_2096) + 1
		for i := range names {
			names[i] = (*C.char)(C.calloc(C.size_t(nameSize), 1))
		}
		columns.ips = (*C.uint32_t)(C.calloc(C.size_t(count), 4))
		columns.classes = (*C.uint8_t)(C.calloc(C.size_t(count), 1))
		var scores [8][]C.uint32_t
		for i := range columns.scores {
			columns.scores[i] = (*C.uint32_t)(C.calloc(C.size_t(count), 4))
			scores[i] = unsafe.Slice(columns.scores[i], count)
		}
		ips, classes := unsafe.Slice(columns.ips, count), unsafe.Slice(columns.classes, count)
		index := 0
		for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
			if player.field_4792 == 0 {
				continue
			}
			info := asPlayerS(player)
			dst := unsafe.Slice((*byte)(unsafe.Pointer(names[index])), nameSize)
			copy(dst, info.Field2096())
			ips[index] = C.uint32_t(bits.ReverseBytes32(nox_xxx_net_getIP_554200(int32(player.playerInd) + 1)))
			classes[index] = C.uint8_t(info.PlayerClass())
			scores[0][index] = player.field_4688
			scores[1][index] = player.field_4688
			scores[2][index] = player.field_4664
			scores[3][index] = player.field_4660
			scores[4][index] = player.field_4668
			scores[5][index] = player.field_4672
			scores[6][index] = player.field_4684
			scores[7][index] = C.uint32_t(sub_4D6540(C.int(player.playerInd)))
			index++
		}
	}
	legacyWords.dword_5d4594_741332 = C.uint32_t(questStatsPlayerCount(header))
}

func buildQuestStatsPacket(columns *C.nox_stats_quest_columns_native, header []byte, length *C.uint32_t) *C.uint16_t {
	var head *nativeStatsField
	fields := [...]struct {
		tag    uintptr
		kind   uint16
		offset int
	}{
		{71944, 6, 4}, {71952, 6, 8}, {71960, 6, 12}, {71968, 7, 24},
		{71976, 7, 280}, {71984, 6, 20}, {71992, 2, 16}, {72000, 3, 0},
	}
	for _, field := range fields {
		tag, data := statsFieldTag(field.tag), header[field.offset:]
		switch field.kind {
		case 6:
			head = statsFieldUint32(head, tag, binary.NativeEndian.Uint32(data))
		case 3:
			head = statsFieldUint16(head, tag, binary.NativeEndian.Uint16(data))
		case 7:
			head = statsPrependField(head, tag, 7, 0, unsafe.Pointer(&data[0]))
		case 2:
			head = statsFieldUint8(head, tag, data[0])
		}
	}
	sequence := memmap.PtrUint32(0x5D4594, 741672)
	head = statsFieldUint32(head, statsFieldTag(72008), *sequence)
	current := memmap.PtrUint32(0x5D4594, 741664)
	*current = 0
	for count := questStatsPlayerCount(header); int(*current) < count; *current++ {
		index := *current
		name := unsafe.Slice(columns.names, count)[index]
		head = statsPrependField(head, statsPlayerTag(71544, index), 7, 0, unsafe.Pointer(name))
		head = statsFieldUint32(head, statsPlayerTag(71552, index), uint32(unsafe.Slice(columns.ips, count)[index]))
		head = statsFieldUint8(head, statsPlayerTag(71560, index), uint8(unsafe.Slice(columns.classes, count)[index]))
		for i, score := range columns.scores {
			tag := statsPlayerTag(71568+8*uintptr(i), index)
			head = statsFieldUint32(head, tag, uint32(unsafe.Slice(score, count)[index]))
		}
	}
	return nox_stats_packet_native_finish(unsafe.Pointer(head), length, (*C.uint32_t)(unsafe.Pointer(sequence)))
}

//export nox_stats_quest_end_native
func nox_stats_quest_end_native() C.int {
	header := questStatsHeader()
	columns := nox_stats_quest_columns_native_get()
	buildQuestStatsPlayers(columns, header)
	length := memmap.PtrUint32(0x5D4594, 741300)
	packet := buildQuestStatsPacket(columns, header, (*C.uint32_t)(unsafe.Pointer(length)))
	return encodeStatsReport(packet, length, 0)
}

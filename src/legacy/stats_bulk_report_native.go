package legacy

/*
#include <stdlib.h>
#include "stats_bulk_native.h"
#include "stats_end_native.h"

typedef const char nox_stats_const_name_t;
typedef const void* nox_stats_const_data_t;

static char* nox_stats_bulk_quest_result(void) { return (char*)(uintptr_t)1; }
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

type nativeStatsField struct {
	name   [4]byte
	kind   uint16
	length uint16
	data   unsafe.Pointer
	next   *nativeStatsField
}

func statsPrependField(head *nativeStatsField, name string, kind, length uint16, data unsafe.Pointer) *nativeStatsField {
	field := (*nativeStatsField)(C.calloc(1, C.size_t(unsafe.Sizeof(nativeStatsField{}))))
	copy(field.name[:], name)
	if kind == 7 {
		length = uint16(alloc.StrLen((*byte)(data)) + 1)
	}
	field.kind, field.length = kind, length
	field.data = C.calloc(1, C.size_t(length))
	copy(unsafe.Slice((*byte)(field.data), int(length)), unsafe.Slice((*byte)(data), int(length)))
	field.next = head
	return field
}

func statsFieldUint8(head *nativeStatsField, name string, value uint8) *nativeStatsField {
	return statsPrependField(head, name, 2, 1, unsafe.Pointer(&value))
}

func statsFieldUint16(head *nativeStatsField, name string, value uint16) *nativeStatsField {
	return statsPrependField(head, name, 3, 2, unsafe.Pointer(&value))
}

func statsFieldUint32(head *nativeStatsField, name string, value uint32) *nativeStatsField {
	return statsPrependField(head, name, 6, 4, unsafe.Pointer(&value))
}

func statsFieldString(head *nativeStatsField, name, value string) *nativeStatsField {
	chars := append([]byte(value), 0)
	return statsPrependField(head, name, 7, 0, unsafe.Pointer(&chars[0]))
}

func statsFreeFields(head *nativeStatsField) {
	for head != nil {
		next := head.next
		C.free(head.data)
		C.free(unsafe.Pointer(head))
		head = next
	}
}

func statsSerializeFields(head *nativeStatsField, length *uint32) *C.uint16_t {
	*length = 4
	for field := head; field != nil; field = field.next {
		*length += uint32(8 + ((int(field.length) + 3) &^ 3))
	}
	packet := C.calloc(1, C.size_t(*length))
	out := unsafe.Slice((*byte)(packet), int(*length))
	binary.BigEndian.PutUint16(out, uint16(*length))
	cursor := out[4:]
	for field := head; field != nil; field = field.next {
		copy(cursor, field.name[:])
		binary.BigEndian.PutUint16(cursor[4:], field.kind)
		binary.BigEndian.PutUint16(cursor[6:], field.length)
		data := unsafe.Slice((*byte)(field.data), int(field.length))
		switch field.kind {
		case 3, 4:
			binary.BigEndian.PutUint16(cursor[8:], binary.NativeEndian.Uint16(data))
		case 5, 6:
			binary.BigEndian.PutUint32(cursor[8:], binary.NativeEndian.Uint32(data))
		default:
			copy(cursor[8:], data)
		}
		cursor = cursor[8+((int(field.length)+3)&^3):]
	}
	return (*C.uint16_t)(packet)
}

type nativeStatsColumns struct {
	names        []string
	ips, teams   []uint32
	classes      []uint8
	active       []uint8
	durations    []uint32
	participants []uint8
	pairs        []byte
}

var statsBulkColumns nativeStatsColumns

func (c *nativeStatsColumns) buildPlayers(header, records []byte, count int) {
	c.names = make([]string, count)
	c.ips = make([]uint32, count)
	c.teams = make([]uint32, count)
	c.classes = make([]uint8, count)
	c.active = make([]uint8, count)
	c.durations = make([]uint32, count)
	c.participants = make([]uint8, count)
	for i := range count {
		field := records[32*i : 32*(i+1)]
		c.names[i] = alloc.GoStringS(field)
		c.ips[i] = binary.NativeEndian.Uint32(field[12:])
		c.teams[i] = binary.NativeEndian.Uint32(field[16:])
		c.classes[i], c.active[i], c.participants[i] = field[20], field[21], field[28]
		elapsed := uint32(time.Now().Unix()) - binary.NativeEndian.Uint32(field[24:])
		binary.NativeEndian.PutUint32(field[24:], elapsed)
		c.durations[i] = elapsed
	}
	binary.NativeEndian.PutUint16(header[6:], uint16(count))
	legacyWords.dword_5d4594_741332 = C.uint32_t(count)
}

func (c *nativeStatsColumns) buildPairs(records []byte, count uint32) {
	c.pairs = append([]byte(nil), records[:2*count]...)
	*memmap.PtrUint32(0x5D4594, 741308) = count
}

func statsFieldTag(offset uintptr) string {
	return alloc.GoStringS(memmap.PtrT[[4]byte](0x587000, offset)[:])
}

func statsPlayerTag(offset uintptr, index uint32) string {
	*memmap.PtrUint8(0x587000, offset+3) = byte(index + '0')
	return statsFieldTag(offset)
}

func statsBuildPacket(columns *nativeStatsColumns, header []byte, mode int, pairs int16, length *uint32) *C.uint16_t {
	var head *nativeStatsField
	fields := [...]struct {
		name   string
		kind   uint16
		offset int
	}{
		{"MXPL", 6, 8}, {"IDNO", 6, 12}, {"GSKU", 6, 16}, {"GSTY", 6, 20}, {"CLGM", 2, 24},
		{"LIMT", 6, 32}, {"TLMT", 6, 36}, {"RSTC", 2, 40}, {"MINE", 6, 44}, {"MAXE", 6, 48},
		{"MINP", 6, 52}, {"MAXP", 6, 56}, {"VIDM", 2, 60}, {"SVRS", 2, 61}, {"NTMS", 2, 25},
		{"SCEN", 7, 96}, {"GNAM", 7, 352}, {"SPL1", 6, 64}, {"SPL2", 6, 68}, {"SPL3", 6, 72},
		{"ARMR", 6, 88}, {"WPN1", 2, 84}, {"WPN2", 2, 85}, {"WPN3", 2, 86}, {"STAF", 6, 92},
		{"DURA", 6, 28},
	}
	for _, field := range fields {
		data := header[field.offset:]
		switch field.kind {
		case 6:
			head = statsFieldUint32(head, field.name, binary.NativeEndian.Uint32(data))
		case 2:
			head = statsFieldUint8(head, field.name, data[0])
		case 7:
			head = statsPrependField(head, field.name, 7, 0, unsafe.Pointer(&data[0]))
		}
	}
	head = statsFieldUint8(head, "FINI", 1)
	head = statsFieldUint8(head, "TRNY", header[26])
	sequence := memmap.PtrUint32(0x5D4594, 741668)
	switch mode {
	case 0:
		*sequence = 0
		head = statsFieldUint32(head, "SEQU", 0)
		head = statsFieldUint8(head, "ENDF", 0)
		head = statsFieldUint16(head, statsFieldTag(71872), 0xffff)
	case 1, 2:
		*sequence++
		head = statsFieldUint32(head, "SEQU", *sequence)
		head = statsFieldUint8(head, "ENDF", uint8(bool2int(mode == 2)))
		playerTag, pairTag := uintptr(71928), uintptr(71936)
		if mode == 2 {
			playerTag, pairTag = 71896, 71904
		}
		count := binary.NativeEndian.Uint16(header[6:])
		head = statsFieldUint16(head, statsFieldTag(playerTag), count)
		current := memmap.PtrUint32(0x5D4594, 741660)
		*current = 0
		for int32(*current) < int32(int16(count)) {
			i := *current
			head = statsFieldString(head, statsPlayerTag(71488, i), columns.names[i])
			head = statsFieldUint32(head, statsPlayerTag(71496, i), columns.ips[i])
			head = statsFieldUint32(head, statsPlayerTag(71512, i), columns.teams[i])
			head = statsFieldUint8(head, statsPlayerTag(71504, i), columns.classes[i])
			head = statsFieldUint8(head, statsPlayerTag(71520, i), columns.active[i])
			head = statsFieldUint32(head, statsPlayerTag(71528, i), columns.durations[i])
			head = statsFieldUint8(head, statsPlayerTag(71536, i), columns.participants[i])
			*current++
		}
		pairLength := uint16(2 * int(pairs))
		data := columns.pairs[:int(pairLength)]
		head = statsPrependField(head, statsFieldTag(pairTag), 20, pairLength, unsafe.Pointer(unsafe.SliceData(data)))
	}
	packet := statsSerializeFields(head, length)
	statsFreeFields(head)
	return packet
}

func statsReport(columns *nativeStatsColumns, header []byte, mode int) C.int {
	length := memmap.PtrUint32(0x5D4594, 741312)
	packet := statsBuildPacket(columns, header, mode, memmap.Int16(0x5D4594, 741308), length)
	return encodeStatsReport(packet, length, C.uint16_t(mode))
}

func encodeStatsReport(packet *C.uint16_t, length *uint32, port C.uint16_t) C.int {
	encoded := sub_42A8B0((*C.uint8_t)(unsafe.Pointer(packet)), (*C.int)(unsafe.Pointer(length)))
	freeStatsPacket(packet)
	var address [72]C.char
	if sub_420360(&address[0], &port) != 0 {
		C.abort()
	}
	freeStatsPacket(encoded)
	return 1
}

func freeStatsPacket(packet *C.uint16_t) {
	C.free(unsafe.Pointer(packet))
}

//export nox_stats_bulk_flush_native
func nox_stats_bulk_flush_native() *C.char {
	if !noxflags.HasGame(noxflags.GameOnline) {
		return nil
	}
	if noxflags.HasGame(noxflags.GameModeQuest) {
		return C.nox_stats_bulk_quest_result()
	}
	header := memmap.Slice(0x5D4594, 599476)
	players := memmap.Slice(0x5D4594, 600124)[:0x2000]
	pairs := memmap.Slice(0x5D4594, 608320)[:0x20000]
	statsBulkColumns.buildPlayers(header, players, int(legacyWords.dword_5d4594_608316))
	statsBulkColumns.buildPairs(pairs, uint32(legacyWords.dword_5d4594_739392))
	binary.NativeEndian.PutUint32(header[28:], uint32(time.Now().Unix())-uint32(legacyWords.dword_5d4594_600116))
	statsReport(&statsBulkColumns, header, 1)

	clear(players)
	clear(pairs)
	legacyWords.dword_5d4594_608316, legacyWords.dword_5d4594_739392 = 0, 0
	for pl := nox_common_playerInfoGetFirst_416EA0(); pl != nil; pl = nox_common_playerInfoGetNext_416EE0(pl) {
		pl.field_4648 = -1
	}
	if host := nox_common_playerInfoFromNum_417090(31); host != nil {
		sub_425F10(host)
	}
	for pl := nox_common_playerInfoGetFirst_416EA0(); pl != nil; pl = nox_common_playerInfoGetNext_416EE0(pl) {
		if pl.playerInd != 31 {
			sub_425F10(pl)
		}
	}
	return nil
}

//export nox_stats_session_report_native
func nox_stats_session_report_native(mode C.int) C.int {
	if mode == 0 {
		statsBulkColumns.names, statsBulkColumns.ips, statsBulkColumns.teams = nil, nil, nil
		statsBulkColumns.classes, statsBulkColumns.pairs = nil, nil
	}
	return statsReport(&statsBulkColumns, memmap.Slice(0x5D4594, 599476), int(mode))
}

//export nox_stats_field_native_prepend
func nox_stats_field_native_prepend(head unsafe.Pointer, name *C.nox_stats_const_name_t, kind, length C.uint16_t, data C.nox_stats_const_data_t) unsafe.Pointer {
	var tag [4]byte
	for i := range tag {
		unit := *(*byte)(unsafe.Add(unsafe.Pointer(name), i))
		if unit == 0 {
			break
		}
		tag[i] = unit
	}
	field := statsPrependField((*nativeStatsField)(head), string(tag[:]), uint16(kind), uint16(length), unsafe.Pointer(data))
	return unsafe.Pointer(field)
}

//export nox_stats_packet_native_finish
func nox_stats_packet_native_finish(head unsafe.Pointer, length, sequence *C.uint32_t) *C.uint16_t {
	fields := (*nativeStatsField)(head)
	packet := statsSerializeFields(fields, (*uint32)(unsafe.Pointer(length)))
	*sequence++
	statsFreeFields(fields)
	return packet
}

package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"bytes"
	"strings"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	mapCycleModes      = 6
	mapCycleMapLimit   = 25
	mapCycleRecordSize = 128
	mapCycleCountOff   = 1548428
	mapCycleIndexOff   = 1548452
	mapCycleEnabledOff = 1548484
	mapCycleEntriesOff = 1529228
	mapCyclePathOff    = 1524108
	mapCyclePathSize   = 1024
)

func mapCycleModeIndex(line string) int32 {
	for i := 0; i < mapCycleModes; i++ {
		p := *memmap.PtrPtr(0x587000, uintptr(191832+8*i))
		if p != nil && strings.EqualFold(alloc.GoString((*byte)(p)), line) {
			return int32(i)
		}
	}
	return -1
}

func mapCycleModeCount(mode int32) *uint32 {
	return memmap.PtrUint32(0x5d4594, mapCycleCountOff+4*uintptr(mode))
}

func appendMapCycleEntry(mode int32, name string) {
	count := mapCycleModeCount(mode)
	offset := mapCycleEntriesOff + mapCycleRecordSize*uintptr(uint32(mode)*mapCycleMapLimit+*count)
	alloc.StrCopyZero(memmap.Slice(0x5d4594, offset)[:mapCycleRecordSize], name)
	*count = *count + 1
}

func readMapCycleLine(file *binfile.File) (string, bool) {
	raw, err := file.ReadString()
	if err != nil {
		return "", false
	}
	// The previous fgets copied at most 126 bytes and accepted only lines
	// completed before EOF. The original trims CR and LF at the first break.
	raw = raw[:min(len(raw), 126)]
	if end := bytes.IndexAny(raw, "\x00\r\n"); end >= 0 {
		raw = raw[:end]
	}
	return string(raw), true
}

//export nox_xxx_loadMapCycle_4D0A30
func nox_xxx_loadMapCycle_4D0A30() *FILE {
	clear(memmap.Slice(0x5d4594, mapCycleCountOff)[:4*mapCycleModes])
	path := memmap.Slice(0x5d4594, mapCyclePathOff)[:mapCyclePathSize]
	alloc.StrCopyZero(path, datapath.Data()+"\\mapcycle.txt")
	input, err := ifs.Open(alloc.GoStringS(path))
	if err != nil {
		*memmap.PtrUint32(0x5d4594, mapCycleEnabledOff) = 0
		return nil
	}
	file := binfile.NewTextFile(input)
	defer file.Close()

	mode := int32(1)
	if line, ok := readMapCycleLine(file); ok {
		if next := mapCycleModeIndex(line); next >= 0 {
			mode = next
		} else {
			appendMapCycleEntry(mode, line)
		}
	}
	for line, ok := readMapCycleLine(file); ok; line, ok = readMapCycleLine(file) {
		if next := mapCycleModeIndex(line); next >= 0 {
			mode = next
			continue
		}
		if line == "" || int32(*mapCycleModeCount(mode)) >= mapCycleMapLimit {
			continue
		}
		mask := sub_4D0C70(C.int(mode))
		name, _, _ := strings.Cut(strings.TrimLeft(line, "."), ".")
		if checkMapFile4CFE10(name) &&
			nox_xxx_mapGetTypeMB_4CFFA0(memmap.PtrOff(0x973F18, 2408))&mask != 0 {
			appendMapCycleEntry(mode, line)
		}
	}
	return nil
}

//export sub_4D0C70
func sub_4D0C70(index C.int) C.int {
	offset := uint32(191836) + 8*uint32(index)
	return C.int(memmap.Uint32(0x587000, uintptr(offset)))
}

//export sub_4D0CC0
func sub_4D0CC0(line *C.char) {
	if line == nil {
		return
	}
	value := GoString(line)
	if index := strings.IndexByte(value, '\r'); index >= 0 {
		*(*byte)(unsafe.Add(unsafe.Pointer(line), index)) = 0
		value = value[:index]
	}
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		*(*byte)(unsafe.Add(unsafe.Pointer(line), index)) = 0
	}
}

//export sub_4D0D50
func sub_4D0D50(mask C.int) C.int {
	for index := C.int(0); index < mapCycleModes; index++ {
		if mask&sub_4D0C70(index) != 0 {
			return index
		}
	}
	return 0
}

//export sub_4D0D90
func sub_4D0D90(value C.int) C.int {
	*memmap.PtrUint32(0x5d4594, mapCycleEnabledOff) = uint32(value)
	return value
}

//export sub_4D0DA0
func sub_4D0DA0() {
	clear(memmap.Slice(0x5d4594, mapCycleIndexOff)[:4*mapCycleModes])
	clear(memmap.Slice(0x5d4594, mapCycleCountOff)[:4*mapCycleModes])
}

//export sub_4D0DC0
func sub_4D0DC0(mask, value C.int) C.int {
	index := sub_4D0D50(mask)
	*memmap.PtrUint32(0x5d4594, mapCycleIndexOff+4*uintptr(index)) = uint32(value)
	return index
}

//export sub_4D0DE0
func sub_4D0DE0(mask C.int) C.int {
	return C.int(memmap.Uint32(0x5d4594, mapCycleIndexOff+4*uintptr(sub_4D0D50(mask))))
}

//export nox_xxx_mapSelectFirst_4D0E00
func nox_xxx_mapSelectFirst_4D0E00() C.int {
	count, _ := questMapFileGlobals4D0F60()
	*count = 0
	suffix := memmap.Slice(0x587000, 192000)[:5]
	for m := nox_common_maplist_first_4D09B0(); m != nil && *count < 128; m = nox_common_maplist_next_4D09C0(m) {
		if m.field_6 == 0 || sub_4CFFC0(m)&C.int(noxflags.GameModeQuest) == 0 {
			continue
		}
		item := (*Nox_map_list_item)(unsafe.Pointer(m))
		entry := questMapEntryNative4D0F60(int32(*count))
		n := copy(entry.Name[:len(entry.Name)-len(suffix)], alloc.GoStringS(item.Name[:]))
		copy(entry.Name[n:], suffix)
		entry.Group = 0
		*count = *count + 1
	}

	nextGroup := uint32(1)
	for index := int32(0); index < int32(*count); index++ {
		entry := questMapEntryNative4D0F60(index)
		if entry.Group != 0 {
			continue
		}
		entry.Group = nextGroup
		nextGroup++
		prefix := alloc.GoStringS(entry.Name[:6])
		for other := index + 1; other < int32(*count); other++ {
			candidate := questMapEntryNative4D0F60(other)
			name := alloc.GoStringS(candidate.Name[:6])
			if len(prefix) == len(name) && questJournalASCIIPrefixEqualFold5007E0(prefix, name, len(prefix)) {
				candidate.Group = entry.Group
			}
		}
	}
	return C.int(*count)
}

//export sub_4D0F30
func sub_4D0F30() {
	count, clock := questMapFileGlobals4D0F60()
	*clock = 1000
	for index := int32(0); index < int32(*count); index++ {
		entry := questMapEntryNative4D0F60(index)
		entry.Uses, entry.LastUsed = 0, 0
	}
}

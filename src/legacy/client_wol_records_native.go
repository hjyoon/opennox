package legacy

/*
#include "GAME1_1.h"
typedef const char global_record_const_char_t;
extern uint32_t dword_5d4594_2660652;
*/
import "C"

import (
	"bytes"
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const wolQueueSize = 23

//export sub_41DA10
func sub_41DA10(index C.int) C.int {
	offset := uintptr(58132 + 16*int(index))
	queue := (*[wolQueueSize]uint16)(*memmap.PtrPtr(0x587000, offset))
	if queue != nil {
		clear(queue[:])
		*memmap.PtrUint32(0x587000, offset+4) = 0
	}
	return 0
}

//export sub_41DA70
func sub_41DA70(index C.int, event C.short) C.int {
	offset := uintptr(58132 + 16*int(index))
	queue := (*[wolQueueSize]uint16)(*memmap.PtrPtr(0x587000, offset))
	if queue != nil {
		count := memmap.PtrUint32(0x587000, offset+4)
		if *count < wolQueueSize {
			queue[*count] = uint16(event)
			(*count)++
			return 16 * index
		}
	}
	result := sub_41E300(11)
	if result != 0 {
		C.dword_5d4594_2660652 = 0
	}
	return C.int(result)
}

//export sub_4207E0
func sub_4207E0() unsafe.Pointer {
	return *memmap.PtrPtr(0x5D4594, 534812)
}

//export nox_xxx_officialStringCmp_41FDE0
func nox_xxx_officialStringCmp_41FDE0() C.int {
	clear(memmap.Slice(0x85B3FC, 10308)[:0x6c])
	var first, second *C.global_record_const_char_t
	if sub_41FBE0(&first, &second) != 1 {
		return 0
	}
	record := sub_4207E0()
	if record == nil {
		return 1
	}
	wolCopyCString(unsafe.Add(record, 228), unsafe.Pointer(first))
	wolCopyCString(unsafe.Add(record, 238), unsafe.Pointer(second))
	sub_41FA50(first)

	const titleOffset = 24
	title := alloc.GoString((*byte)(unsafe.Add(record, titleOffset)))
	separator := strings.IndexByte(title, ':')
	if separator < 0 {
		separator = len(title)
	}
	const official = "Official"
	prefix := unsafe.Slice((*byte)(unsafe.Add(record, titleOffset+separator+1)), len(official))
	if bytes.Equal(prefix, []byte(official)) {
		noxflags.SetGame(noxflags.GameFlag25)
	} else {
		noxflags.UnsetGame(noxflags.GameFlag25)
	}
	return 1
}

func wolCopyCString(destination, source unsafe.Pointer) {
	src := (*byte)(source)
	size := alloc.StrLen(src) + 1
	copy(unsafe.Slice((*byte)(destination), size), unsafe.Slice(src, size))
}

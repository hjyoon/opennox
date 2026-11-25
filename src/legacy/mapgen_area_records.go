package legacy

/*
#include "GAME4_2.h"
#include "cgo_const.h"

extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487676;
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const mapgenAreaRecordBytes = 64

var _ [mapgenAreaRecordBytes]byte = [unsafe.Sizeof(mapGenDecorAreaFlags{})]byte{}

func mapgenAreaRecords() *mapGenDecorAreaFlags {
	return (*mapGenDecorAreaFlags)(resolveMapgenPointer(uintptr(C.dword_5d4594_2487672)))
}

func mapgenAreaRecordName(record *mapGenDecorAreaFlags) []byte {
	// The archive copies the whole 64-byte name slot. Its last word is
	// also where the generator stores attachment flags.
	return unsafe.Slice((*byte)(unsafe.Pointer(record)), mapgenAreaRecordBytes)
}

func mapgenAreaRecordAt(index int32) *mapGenDecorAreaFlags {
	records := mapgenAreaRecords()
	if records == nil || index < 0 || index > int32(C.dword_5d4594_2487676) {
		return nil
	}
	return (*mapGenDecorAreaFlags)(unsafe.Add(unsafe.Pointer(records), uintptr(index)*mapgenAreaRecordBytes))
}

func mapgenFindAreaRecord(name string) int32 {
	count, records := int32(C.dword_5d4594_2487676), mapgenAreaRecords()
	if count <= 0 || records == nil {
		return -1
	}
	entries := unsafe.Slice(records, int(count))
	for i := range entries {
		if alloc.GoStringS(mapgenAreaRecordName(&entries[i])) == name {
			return int32(i)
		}
	}
	return -1
}

//export sub_5268F0
func sub_5268F0(name *C.nox_cgo_const_char) C.int {
	if name == nil {
		return -1
	}
	return C.int(mapgenFindAreaRecord(GoStringP(unsafe.Pointer(name))))
}

func mapgenAreaAttachmentFlags(name string) uint32 {
	pieces := strings.FieldsFunc(name, func(r rune) bool { return r == '-' })
	if len(pieces) < 3 {
		return 0
	}
	var flags uint32
	for i := uintptr(0); i < 32; i++ {
		key := *memmap.PtrUint8(0x587000, 255032+i)
		if key == 0 {
			break
		}
		if strings.IndexByte(pieces[2], key) >= 0 {
			flags |= 1 << i
		}
	}
	return flags
}

//export nox_mapgenLoadAreaRecordsNative
func nox_mapgenLoadAreaRecordsNative() C.uintptr_t {
	sub_526A90()
	count := sub_502A20()
	if count < 0 {
		C.dword_5d4594_2487676 = 0
		return 0
	}
	ptr := nox_mapgen_alloc_memory(C.size_t(count), mapgenAreaRecordBytes)
	if ptr == nil {
		C.dword_5d4594_2487676 = 0
		return 0
	}
	token := registerMapgenPointer(ptr)
	if token == 0 {
		nox_mapgen_release_memory(ptr)
		C.dword_5d4594_2487676 = 0
		return 0
	}
	records := unsafe.Slice((*mapGenDecorAreaFlags)(ptr), int(count))
	for i := range records {
		name := GoString(sub_5029F0(C.int(i)))
		record := &records[i]
		if !copyMapgenName(mapgenAreaRecordName(record), name) {
			mapgenReleaseDecorRecord(ptr)
			C.dword_5d4594_2487676 = 0
			return 0
		}
		record.flags |= mapgenAreaAttachmentFlags(name)
	}
	C.dword_5d4594_2487672 = C.uint32_t(token)
	C.dword_5d4594_2487676 = C.uint32_t(count)
	if count == 0 {
		return C.uintptr_t(uintptr(ptr))
	}
	return C.uintptr_t(count)
}

//export sub_526AA0
func sub_526AA0(index C.int) *C.char {
	return (*C.char)(unsafe.Pointer(mapgenAreaRecordAt(int32(index))))
}

package legacy

/*
#include <stdint.h>
extern unsigned char byte_581450[23472];
extern unsigned char byte_587000[316820];
extern unsigned char byte_5D4594[2598284];
extern unsigned char byte_973CE0[568];
extern unsigned char byte_973F18[44881];
extern unsigned char byte_85B3FC[1029636];
extern unsigned char byte_852978[40];
extern unsigned char byte_973A20[704];
*/
import "C"
import (
	"slices"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func init() {
	memmap.RegisterBlobData(0x581450, "byte_581450", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_581450[0])), int(unsafe.Sizeof(C.byte_581450))))
	memmap.RegisterBlobData(0x587000, "byte_587000", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_587000[0])), int(unsafe.Sizeof(C.byte_587000))))
	memmap.RegisterBlobData(0x5D4594, "byte_5D4594", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_5D4594[0])), int(unsafe.Sizeof(C.byte_5D4594))))
	memmap.RegisterBlobData(0x973CE0, "byte_973CE0", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_973CE0[0])), int(unsafe.Sizeof(C.byte_973CE0))))
	memmap.RegisterBlobData(0x973F18, "byte_973F18", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_973F18[0])), int(unsafe.Sizeof(C.byte_973F18))))
	memmap.RegisterBlobData(0x85B3FC, "byte_85B3FC", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_85B3FC[0])), int(unsafe.Sizeof(C.byte_85B3FC))))
	memmap.RegisterBlobData(0x852978, "byte_852978", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_852978[0])), int(unsafe.Sizeof(C.byte_852978))))
	memmap.RegisterBlobData(0x973A20, "byte_973A20", unsafe.Slice((*byte)(unsafe.Pointer(&C.byte_973A20[0])), int(unsafe.Sizeof(C.byte_973A20))))
}

//export mem_getPtrValue
func mem_getPtrValue(base, off C.uintptr_t) unsafe.Pointer {
	return *memmap.PtrPtr(uintptr(base), uintptr(off))
}

//export mem_setPtrValue
func mem_setPtrValue(base, off C.uintptr_t, value unsafe.Pointer) {
	*memmap.PtrPtr(uintptr(base), uintptr(off)) = value
}

// Keep the original switch order for lookups that fall through their first blob.
var legacyMemoryBlobOrder = [...]uintptr{
	0x581450, 0x587000, 0x5D4594, 0x973CE0, 0x973F18, 0x85B3FC, 0x852978, 0x973A20,
}

func legacyMemoryBlobPointer(blob *memmap.Blob, off, size uintptr) unsafe.Pointer {
	if blob == nil || len(blob.Data) == 0 {
		return nil
	}
	capacity := uintptr(len(blob.Data))
	if off > capacity || size > capacity-off {
		return nil
	}
	return unsafe.Add(unsafe.Pointer(&blob.Data[0]), off)
}

func legacyMemoryPointer(base, off, size uintptr) unsafe.Pointer {
	blob := memmap.BlobByAddr(base)
	if blob == nil || blob.Addr != base {
		return nil
	}
	if pointer := legacyMemoryBlobPointer(blob, off, size); pointer != nil {
		return pointer
	}
	next := slices.Index(legacyMemoryBlobOrder[:], base)
	if next < 0 {
		return nil
	}
	for _, address := range legacyMemoryBlobOrder[next+1:] {
		if pointer := legacyMemoryBlobPointer(memmap.BlobByAddr(address), off, size); pointer != nil {
			return pointer
		}
	}
	return nil
}

//export nox_legacy_memory_pointer
func nox_legacy_memory_pointer(base, off, size C.uintptr_t) unsafe.Pointer {
	return legacyMemoryPointer(uintptr(base), uintptr(off), uintptr(size))
}

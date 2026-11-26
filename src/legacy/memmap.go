package legacy

/*
#include "memmap.h"
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

//export mem_getPtr
func mem_getPtr(base, off C.uintptr_t) unsafe.Pointer {
	return legacyCMapPointer(uintptr(base), uintptr(off), unsafe.Sizeof(C.int(0)))
}

//export mem_getPtrSize
func mem_getPtrSize(base, off, size C.uintptr_t) unsafe.Pointer {
	return legacyCMapPointer(uintptr(base), uintptr(off), uintptr(size))
}

//export mem_getU8Ptr
func mem_getU8Ptr(base, off C.uintptr_t) *C.uint8_t {
	return (*C.uint8_t)(legacyCMapPointer(uintptr(base), uintptr(off), 1))
}

//export mem_getI8Ptr
func mem_getI8Ptr(base, off C.uintptr_t) *C.int8_t {
	return (*C.int8_t)(legacyCMapPointer(uintptr(base), uintptr(off), 1))
}

//export mem_getU16Ptr
func mem_getU16Ptr(base, off C.uintptr_t) *C.uint16_t {
	return (*C.uint16_t)(legacyCMapPointer(uintptr(base), uintptr(off), 2))
}

//export mem_getI16Ptr
func mem_getI16Ptr(base, off C.uintptr_t) *C.int16_t {
	return (*C.int16_t)(legacyCMapPointer(uintptr(base), uintptr(off), 2))
}

//export mem_getU32Ptr
func mem_getU32Ptr(base, off C.uintptr_t) *C.uint32_t {
	return (*C.uint32_t)(legacyCMapPointer(uintptr(base), uintptr(off), 4))
}

//export mem_getI32Ptr
func mem_getI32Ptr(base, off C.uintptr_t) *C.int32_t {
	return (*C.int32_t)(legacyCMapPointer(uintptr(base), uintptr(off), 4))
}

//export mem_getU64Ptr
func mem_getU64Ptr(base, off C.uintptr_t) *C.uint64_t {
	return (*C.uint64_t)(legacyCMapPointer(uintptr(base), uintptr(off), 8))
}

//export mem_getI64Ptr
func mem_getI64Ptr(base, off C.uintptr_t) *C.int64_t {
	return (*C.int64_t)(legacyCMapPointer(uintptr(base), uintptr(off), 8))
}

//export mem_getFloatPtr
func mem_getFloatPtr(base, off C.uintptr_t) *C.float {
	return (*C.float)(legacyCMapPointer(uintptr(base), uintptr(off), 4))
}

//export mem_getDoublePtr
func mem_getDoublePtr(base, off C.uintptr_t) *C.double {
	return (*C.double)(legacyCMapPointer(uintptr(base), uintptr(off), 8))
}

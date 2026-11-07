package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func sub40A180Native(value int16) uint8 {
	index := sub409A70Native(value)
	return *(*uint8)(memmap.PtrOff(0x5D4594, 3500+uintptr(index)))
}

func sub40A220Native() int32 {
	return int32(*(*uint32)(memmap.PtrOff(0x587000, 4660)))
}

func sub40A3C0Native(value uint32) uint32 {
	if value > 0x800 && value < 0x8000 {
		*(*uint32)(memmap.PtrOff(0x587000, 4652)) = value
	}
	return value
}

func sub40A6A0Native(value int32) int32 {
	*(*uint32)(memmap.PtrOff(0x5D4594, 3588)) = uint32(value)
	return value
}

func sub40A6B0Native() int32 {
	return int32(*(*uint32)(memmap.PtrOff(0x5D4594, 3588)))
}

// The sealed original checks five live key/value pairs in order. A duplicate
// key returns the first payload; an absent key returns the raw DWORD 1.
func sub40A710TableNative(value int32, table unsafe.Pointer) int32 {
	for index := 0; index < 5; index++ {
		if *(*uint32)(unsafe.Add(table, 8*index)) == uint32(value) {
			result := *(*uint32)(unsafe.Add(table, 8*index+4))
			runtime.KeepAlive(table)
			return int32(result)
		}
	}
	runtime.KeepAlive(table)
	return 1
}

func sub40A710Native(value int32) int32 {
	return sub40A710TableNative(value, memmap.PtrOff(0x587000, 4664))
}

package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type namedBitmaskNames4239C0 func(uint32) unsafe.Pointer

func namedBitmaskNativeNames4239C0(names unsafe.Pointer) namedBitmaskNames4239C0 {
	return func(index uint32) unsafe.Pointer {
		return *(*unsafe.Pointer)(unsafe.Add(names, uintptr(index)*unsafe.Sizeof(uintptr(0))))
	}
}

func namedBitmaskMonsterNames4239C0(index uint32) unsafe.Pointer {
	// The monster definition table is packed PE32 storage, not a native char**.
	return *memmap.PtrPtr(0x587000, 247536+4*uintptr(index))
}

func namedBitmaskSetNative4239C0(name unsafe.Pointer, bitmask *uint32, names namedBitmaskNames4239C0) int32 {
	for index := uint32(0); ; index++ {
		allowed := names(index)
		if allowed == nil {
			return 0
		}
		if namedBitmaskCompare4239C0(allowed, name) == 0 {
			// Original EDI is32bit, and x86 SHL uses only its low five bits.
			*bitmask |= uint32(1) << (index & 31)
			return 1
		}
	}
}

func namedBitmaskParseNative423930(input unsafe.Pointer, bitmask *uint32, names namedBitmaskNames4239C0) {
	// Preserve the original libc tokenizer state, including CRT implementations
	// that keep strtok's continuation in thread-local storage. Its retained
	// pointer must refer to foreign storage across Cgo calls.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	text := alloc.GoString((*byte)(input))
	copyBuf, free := alloc.Make([]byte{}, len(text)+1)
	defer free()
	copy(copyBuf, text)
	scratch := unsafe.Pointer(&copyBuf[0])
	if namedBitmaskPrefix423930(scratch) == 0 {
		return
	}
	for token := namedBitmaskToken423930(scratch); token != nil; token = namedBitmaskToken423930(nil) {
		namedBitmaskSetNative4239C0(token, bitmask, names)
	}
}

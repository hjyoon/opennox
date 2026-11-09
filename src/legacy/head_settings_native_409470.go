package legacy

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

// Pointers always originate in a real owner. The packed settings storage is
// unchanged; the two independent C globals are not memmap shadows or a struct.
func headSettingsCell409470(base, off uintptr) unsafe.Pointer { return memmap.PtrOff(base, off) }
func headSettingsDirty409470(v int32)                         { *(*int32)(settingFlagsNativeCell409E40(1)) = v }
func headSettingsDirtyGet409470() int32                       { return *(*int32)(settingFlagsNativeCell409E40(1)) }
func headSettingsPath409470() unsafe.Pointer                  { return headSettingsCell409470(0x5d4594, 2598188) }
func headSettingsMap409470() unsafe.Pointer                   { return headSettingsCell409470(0x85b3fc, 36) }
func headSettingsName409470() unsafe.Pointer                  { return headSettingsCell409470(0x5d4594, 1324) }
func headSettingsPass409470() unsafe.Pointer                  { return headSettingsCell409470(0x5d4594, 3540) }

// This ABI supplies no destination length. As in the old C entry, the caller
// owns sufficient token storage. EOF leaves the exact partial bytes unclosed.
// Read's count is deliberately ignored; LastErr is checked immediately after
// every read, and the previous byte survives whitespace and skipLine calls.
func headSettingsParse409470(file, out unsafe.Pointer) int32 {
	var word uint32
	var current uint32
	off := uintptr(0)
	leading := true
	for {
		previous := current
		headSettingsRead409470(file, unsafe.Pointer(&word))
		if headSettingsLastErr409470(file) == -1 {
			runtime.KeepAlive(out)
			return 0
		}
		current = word
		if headSettingsSpace409470(uint16(word)) {
			if leading {
				continue
			}
			*(*byte)(unsafe.Add(out, off)) = 0
			runtime.KeepAlive(file)
			runtime.KeepAlive(out)
			return 1
		}
		leading = false
		if current != '/' || previous != '/' {
			*(*byte)(unsafe.Add(out, off)) = byte(current)
			off++
		} else {
			headSettingsSkipLine409470(file)
			off = 0
			current = word
			leading = true
		}
	}
}

func headSettingsCStringLen409470(p unsafe.Pointer) uintptr {
	var n uintptr
	for *(*byte)(unsafe.Add(p, n)) != 0 {
		n++
	}
	runtime.KeepAlive(p)
	return n
}
func headSettingsStrncpy409470(dst, src unsafe.Pointer, n uintptr) {
	ended := false
	for i := uintptr(0); i < n; i++ {
		var b byte
		if !ended {
			b = *(*byte)(unsafe.Add(src, i))
			ended = b == 0
		}
		*(*byte)(unsafe.Add(dst, i)) = b
	}
	runtime.KeepAlive(src)
	runtime.KeepAlive(dst)
}

func headSettingsSetPath409470(src unsafe.Pointer) unsafe.Pointer {
	// Native safety policy only: nil is a no-op; a basename shorter than the
	// four-byte extension becomes empty; oversized basenames are clipped to
	// 79 bytes before the next registered memmap field at offset 116. The old
	// normal domain (non-nil, basename length 4..83 inclusive) is unchanged.
	if src == nil {
		return nil
	}
	dst := headSettingsPath409470()
	if headSettingsCompare409470(dst, src) == 0 {
		runtime.KeepAlive(src)
		return nil
	}
	headSettingsStrncpy409470(dst, src, 80)
	*(*byte)(unsafe.Add(dst, 79)) = 0
	var suffix unsafe.Pointer
	for i := uintptr(0); ; i++ {
		b := *(*byte)(unsafe.Add(src, i))
		if b == 0 {
			break
		}
		if b == '\\' {
			suffix = unsafe.Add(src, i+1)
		}
	}
	if suffix == nil {
		suffix = dst
	}
	n := headSettingsCStringLen409470(suffix)
	if n >= 4 {
		n -= 4
	} else {
		n = 0
	}
	if n > 79 {
		n = 79
	}
	name := headSettingsMap409470()
	headSettingsStrncpy409470(name, suffix, n)
	*(*byte)(unsafe.Add(name, n)) = 0
	headSettingsDirty409470(1)
	runtime.KeepAlive(src)
	return name
}

func headSettingsSubflags409470() int32 { return int32(*(*uint32)(settingFlagsNativeCell409E40(0))) }
func headSettingsSetLimit409470(v int32) int32 {
	p := (*uint32)(headSettingsCell409470(0x5d4594, 3464))
	if *p != uint32(v) {
		*p = uint32(v)
		headSettingsDirty409470(1)
	}
	return v
}
func headSettingsGetLimit409470() int32 {
	return int32(*(*uint32)(headSettingsCell409470(0x5d4594, 3464)))
}
func headSettingsGamedata409470(v int16) int16 {
	index := sub409A70Native(v)
	return int16(*(*uint16)(headSettingsCell409470(0x5d4594, uintptr(3488+2*index))))
}
func headSettingsSet3512409470(v int32) {
	*(*uint32)(headSettingsCell409470(0x5d4594, 3512)) = uint32(v)
}
func headSettingsGet3512409470() int32 {
	return int32(*(*uint32)(headSettingsCell409470(0x5d4594, 3512)))
}

func headSettingsSetName409470(src unsafe.Pointer) unsafe.Pointer {
	dst := headSettingsName409470()
	if src == nil {
		*(*byte)(dst) = 0
		headSettingsDirty409470(1)
		return nil
	}
	var tmp [16]byte
	headSettingsStrncpy409470(unsafe.Pointer(&tmp[0]), src, 15)
	if headSettingsCompare409470(dst, unsafe.Pointer(&tmp[0])) == 0 {
		runtime.KeepAlive(src)
		return nil
	}
	headSettingsStrncpy409470(dst, unsafe.Pointer(&tmp[0]), 15)
	*(*byte)(unsafe.Add(dst, 15)) = 0
	headSettingsDirty409470(1)
	runtime.KeepAlive(src)
	return dst
}
func headSettingsSetRespawn409470(v int32) int32 {
	*(*uint32)(headSettingsCell409470(0x5d4594, 3584)) = uint32(v)
	return v
}
func headSettingsRespawn409470() int32 {
	if headSettingsFlags409470(4096) {
		return 0
	}
	return int32(*(*uint32)(headSettingsCell409470(0x5d4594, 3584)))
}
func headSettingsSetPass409470(src unsafe.Pointer) unsafe.Pointer {
	dst := headSettingsPass409470()
	// Public nox_wcscpy ABI has no capacity. Preserve its caller-capacity
	// contract and forward copy order; do not infer a new struct or widen it.
	for off := uintptr(0); ; off += 2 {
		v := *(*uint16)(unsafe.Add(src, off))
		*(*uint16)(unsafe.Add(dst, off)) = v
		if v == 0 {
			break
		}
	}
	runtime.KeepAlive(src)
	return dst
}
func headSettingsRate409470() int32 { return int32(*(*uint32)(headSettingsCell409470(0x587000, 4728))) }
func headSettingsRateUpdate409470(v int32) int32 {
	p := (*uint32)(headSettingsCell409470(0x587000, 4728))
	result := int32(*p)
	if uint32(v) != *p {
		result = 0
		if headSettingsFlags409470(0x20000) {
			result = headSettingsNotify409470(159)
		}
	}
	// In particular, callbacks see the OLD rate and their mutation is then
	// overwritten by the supplied DWORD exactly as in the prior C body.
	*p = uint32(v)
	return result
}

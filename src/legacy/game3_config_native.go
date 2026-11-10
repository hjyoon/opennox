package legacy

/*
#include <stdlib.h>
#include <string.h>
#include <wctype.h>
#include "compat.h"
#include "GAME3.h"
#include "game3_class_config_options.h"
extern uint32_t dword_587000_171388;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func game3ConfigCopy(s *C.char, off uintptr, limit int) C.int {
	if s == nil {
		return 0
	}
	text := GoString(s)
	if limit >= 0 && len(text) > limit {
		return 0
	}
	dst := memmap.Slice(0x5D4594, off)
	copy(dst[:len(text)+1], text)
	dst[len(text)] = 0
	return 1
}

//export game3_config_copy_4A7A80
func game3_config_copy_4A7A80(s *C.char) C.int { return game3ConfigCopy(s, 1308644, -1) }

//export game3_config_copy_4A7AC0
func game3_config_copy_4A7AC0(s *C.char) C.int { return game3ConfigCopy(s, 1308172, -1) }

//export game3_config_copy_4A7B00
func game3_config_copy_4A7B00(s *C.char) C.int { return game3ConfigCopy(s, 1308352, -1) }

//export game3_config_copy_4A7BC0
func game3_config_copy_4A7BC0(s *C.char) C.int { return game3ConfigCopy(s, 1308324, -1) }

//export game3_config_copy_4A7C00
func game3_config_copy_4A7C00(s *C.char) C.int { return game3ConfigCopy(s, 1308364, -1) }

//export game3_config_copy_4A7D00
func game3_config_copy_4A7D00(s *C.char) C.int { return game3ConfigCopy(s, 1308192, 128) }

//export sub_4A7A60
func sub_4A7A60(v C.int) C.int { C.dword_587000_171388 = C.uint32_t(v); return v }

//export sub_4A7A70
func sub_4A7A70(v C.int) C.int { *memmap.PtrUint32(0x5D4594, 1308168) = uint32(v); return v }

//export sub_4A7B40
func sub_4A7B40(s *C.char) C.int {
	if s == nil {
		return 0
	}
	for i := uintptr(0); ; i++ {
		name := *memmap.PtrPtr(0x587000, 171856+8*i)
		if name == nil {
			return 1
		}
		if C.nox_strcmpi((*C.char)(name), s) == 0 {
			*memmap.PtrUint32(0x5D4594, 1308184) = *memmap.PtrUint32(0x587000, 171860+8*i)
			return 1
		}
	}
}

func game3ConfigAtoi(s *C.char, off uintptr) C.int {
	if s == nil {
		return 0
	}
	*memmap.PtrUint32(0x5D4594, off) = uint32(C.atoi(s))
	return 1
}

//export sub_4A7BA0
func sub_4A7BA0(s *C.char) C.int { return game3ConfigAtoi(s, 1308740) }

//export sub_4A7C40
func sub_4A7C40(s *C.char) C.int { return game3ConfigAtoi(s, 1308188) }

//export sub_4A7CE0
func sub_4A7CE0(s *C.char) C.int { return game3ConfigAtoi(s, 1308728) }

//export sub_4A7D50
func sub_4A7D50(s *C.char) C.int { return game3ConfigAtoi(s, 1308348) }

//export sub_4A7C60
func sub_4A7C60(s *C.char) C.int {
	if s == nil {
		return 0
	}
	*memmap.PtrUint32(0x5D4594, 1308736) = 0
	*memmap.PtrUint32(0x5D4594, 1308732) = 0
	// Keep the host CRT tokenizer cursor and the original unequal delimiter
	// sets: the second token permits commas and the first permits spaces.
	if p := C.strtok(s, internCStr(",\t\r\n")); p != nil {
		*memmap.PtrUint32(0x5D4594, 1308732) = uint32(C.atoi(p))
	}
	if p := C.strtok(nil, internCStr(" \t\r\n")); p != nil {
		*memmap.PtrUint32(0x5D4594, 1308736) = uint32(C.atoi(p))
	}
	if *memmap.PtrUint32(0x5D4594, 1308732) != 0 && *memmap.PtrUint32(0x5D4594, 1308736) != 0 {
		return 1
	}
	return 0
}

//export sub_4A7EF0
func sub_4A7EF0() *C.char { return (*C.char)(memmap.PtrOff(0x5D4594, 1308732)) }

//export sub_4A6B50
func sub_4A6B50(p *C.wchar2_t) C.int {
	// Valid NUL-terminated caller memory is required, as before. A Go scratch
	// slice additionally avoids the old 26-word local overflow for long names.
	n := 0
	for *(*uint16)(unsafe.Add(unsafe.Pointer(p), 2*n)) != 0 {
		n++
	}
	s := unsafe.Slice((*uint16)(unsafe.Pointer(p)), n+1)
	space := func(v uint16) bool { return C.iswspace(C.wint_t(v)) != 0 }
	first := 0
	for first < n && space(s[first]) {
		first++
	}
	if first == n {
		return 0
	}
	switch s[first] {
	case '*', '?', '<', '>', '\\', '/', ':', '"', '|':
		s[first] = '-'
	}
	tmp := append([]uint16(nil), s[first:n]...)
	copy(s, tmp)
	s[len(tmp)] = 0
	last := len(tmp) - 1
	for last >= 0 && space(tmp[last]) {
		last--
	}
	s[last+1] = 0
	return 1
}

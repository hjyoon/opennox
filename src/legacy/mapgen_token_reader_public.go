package legacy

/*
#include <ctype.h>
#include "GAME4_2.h"

static inline int nox_mapgen_token_isspace(unsigned char value) {
	return isspace((int)value);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func mapGenCommentLine51E630(read func(*byte) int32, nextLine func()) int32 {
	for {
		var value byte
		status := read(&value)
		if status == -1 {
			return status
		}
		if value == '\n' {
			nextLine()
			return status
		}
	}
}

func mapGenPlainToken51E570(read func(*byte) int32, put func(int, byte), isSpace func(byte) bool, nextLine func()) int32 {
	var value byte
	leading := true
	index := 0
	for {
		previous := value
		if read(&value) == -1 {
			return 0
		}
		if value == '\n' {
			nextLine()
		}
		if isSpace(value) {
			if !leading {
				put(index, 0)
				return 1
			}
			continue
		}
		if previous == '/' && value == '/' {
			mapGenCommentLine51E630(read, nextLine)
			index, leading = 0, true
			continue
		}
		put(index, value)
		index++
		leading = false
	}
}

func mapGenDirectiveToken51E540(plain func() int32, directive func() int32) int32 {
	if plain() == 0 {
		return 0
	}
	return int32(bool2int(directive() != 0))
}

func mapGenTokenReadByte(file *C.FILE, value *byte) int32 {
	nox_binfile_fread_408E40((*C.char)(unsafe.Pointer(value)), 1, 1, file)
	return nox_binfile_lastErr_409370(file)
}

func mapGenTokenNextLine() {
	ptr := memmap.PtrUint32(0x5D4594, 2487520)
	*ptr++
}

//export sub_51E630
func sub_51E630(file *C.FILE) C.int {
	return C.int(mapGenCommentLine51E630(func(value *byte) int32 {
		return mapGenTokenReadByte(file, value)
	}, mapGenTokenNextLine))
}

//export sub_51E570
func sub_51E570(file *C.FILE, dest *C.uint8_t) C.int {
	return mapgenReadPlainToken(file, func(index int, value byte) {
		*(*byte)(unsafe.Add(unsafe.Pointer(dest), index)) = value
	})
}

func mapgenReadPlainToken(file *C.FILE, put func(int, byte)) C.int {
	return C.int(mapGenPlainToken51E570(
		func(value *byte) int32 { return mapGenTokenReadByte(file, value) },
		put,
		func(value byte) bool { return C.nox_mapgen_token_isspace(C.uchar(value)) != 0 },
		mapGenTokenNextLine,
	))
}

//export nox_xxx_mapGenReadLine_51E540
func nox_xxx_mapGenReadLine_51E540(file *C.FILE, dest *C.uint8_t) C.int {
	return C.int(mapGenDirectiveToken51E540(
		func() int32 { return int32(sub_51E570(file, dest)) },
		func() int32 { return int32(sub_51E670(file)) },
	))
}

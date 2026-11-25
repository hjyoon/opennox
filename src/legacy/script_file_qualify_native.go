package legacy

/*
#include <stdio.h>
#include "cgo_const.h"
#include "server__script__file.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func qualifyScriptFileFunctions(source, destination *FILE, count C.int, ordinal, x, y C.int32_t) bool {
	// The original qualifier reads these names into a 2 KiB array.
	var name [2048]byte
	for i := C.int(0); i < count; i++ {
		if !transferScriptBytes(source, destination, 4) {
			return false
		}
		size, ok := readScriptInt(source)
		if !ok || size < 0 || size >= C.int(len(name)) {
			return false
		}
		if nox_fs_fread(source, unsafe.Pointer(&name[0]), int32(size)) != int32(size) {
			return false
		}
		name[size] = 0
		payload := name[:size]
		if i > 1 {
			qualified := sub_5435C0((*C.nox_cgo_const_char)(unsafe.Pointer(&name[0])), ordinal, x, y)
			payload = []byte(GoString(qualified))
		}
		nox_script_writeInt_542BB0(C.int(len(payload)), destination)
		if len(payload) > 0 && nox_fs_fwrite(destination, unsafe.Pointer(&payload[0]), int32(len(payload))) != int32(len(payload)) {
			return false
		}
		if !transferScriptBytes(source, destination, 8) {
			return false
		}
		nox_script_readWriteIii_541D80(source, destination)
		if !transferScriptBytes(source, destination, 4) {
			return false
		}
		lengthPos := nox_fs_ftell(destination)
		if lengthPos < 0 || !transferScriptBytes(source, destination, 4) {
			return false
		}
		if nox_script_readWriteYyy_542380(source, destination, 0) != 1 {
			return false
		}
		length := nox_fs_ftell(destination) - lengthPos - 4
		nox_fs_fseek(destination, lengthPos, C.SEEK_SET)
		nox_script_writeInt_542BB0(C.int(length), destination)
		nox_fs_fseek(destination, length, C.SEEK_CUR)
	}
	return true
}

func qualifyScriptFileContents(source, destination *FILE, ordinal, x, y C.int32_t) bool {
	if !transferScriptBytes(source, destination, 12) {
		return false
	}
	stringCount, ok := readScriptInt(source)
	if !ok || stringCount < 0 {
		return false
	}
	nox_script_writeInt_542BB0(stringCount, destination)
	for range stringCount {
		if !transferLengthPrefixedScriptBytes(source, destination) {
			return false
		}
	}
	if !transferScriptBytes(source, destination, 4) {
		return false
	}
	count, ok := readScriptInt(source)
	if !ok || count < 0 {
		return false
	}
	nox_script_writeInt_542BB0(count, destination)
	return qualifyScriptFileFunctions(source, destination, count, ordinal, x, y) &&
		transferScriptBytes(source, destination, 4)
}

//export sub_543110
func sub_543110(sourceName *C.nox_cgo_const_char, offset *C.int) C.int {
	xy := (*[2]C.int)(unsafe.Pointer(offset))
	ordinal := C.int32_t(Get_dword_5d4594_3835312())
	x, y := C.int32_t(xy[0]), C.int32_t(xy[1])
	sub_542BF0(ordinal, x, y)

	prefix := alloc.GoStringS(memmap.Slice(0x973F18, 42152))
	suffix := alloc.GoStringS(memmap.Slice(0x587000, 282608)[:12])
	temporary := CString(prefix + suffix)
	defer StrFree(temporary)
	if !bool(nox_fs_copy(sourceName, temporary)) {
		return 0
	}
	moved := false
	defer func() {
		if !moved {
			nox_fs_remove(temporary)
		}
	}()
	source := nox_fs_open(sourceName)
	if source == nil {
		return 0
	}
	destination := nox_fs_open_rw(temporary)
	if destination == nil {
		nox_fs_close(source)
		return 0
	}
	ok := qualifyScriptFileContents(source, destination, ordinal, x, y)
	nox_fs_close(source)
	nox_fs_close(destination)
	if !ok {
		return 0
	}
	moved = bool(nox_fs_move(temporary, sourceName))
	return C.int(bool2int(moved))
}

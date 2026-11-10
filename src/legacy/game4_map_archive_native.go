package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME4.h"
#include "common/fs/nox_fs.h"

extern char* dword_5d4594_1599576;
extern void* dword_5d4594_1599588;
extern void* dword_5d4594_1599592;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_1599644;
extern uint32_t dword_5d4594_3835396;
extern FILE* nox_file_8;

*/
import "C"

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const game4MapRecordBytes = 76

func game4MapRecord(index int32) unsafe.Pointer {
	return unsafe.Add(unsafe.Pointer(C.dword_5d4594_1599576), uintptr(index)*game4MapRecordBytes)
}

//export sub_5029A0
func sub_5029A0(name *C.char) C.int {
	if int32(C.dword_5d4594_1599596) <= 0 {
		return -1
	}
	for i := int32(0); ; {
		if C.nox_strcmpi(name, (*C.char)(game4MapRecord(i))) == 0 {
			return C.int(i)
		}
		i++
		if i >= int32(C.dword_5d4594_1599596) {
			return -1
		}
	}
}

//export sub_5029F0
func sub_5029F0(index C.int) *C.char {
	// The inclusive upper bound is the current C contract (one-past record).
	if index < 0 || int32(index) > int32(C.dword_5d4594_1599596) {
		return nil
	}
	return (*C.char)(game4MapRecord(int32(index)))
}

//export sub_502A20
func sub_502A20() C.int { return C.int(int32(C.dword_5d4594_1599596)) }

//export sub_502A30
func sub_502A30(name *C.char) C.int { return sub_502D70(sub_5029A0(name)) }

//export sub_502A50
func sub_502A50(name *C.char) C.int {
	sub_502DF0()
	if name != nil {
		C.strncpy((*C.char)(C.dword_5d4594_1599588), name, 0x7ff)
		return 1
	}
	*(*byte)(C.dword_5d4594_1599588) = *memmap.PtrUint8(0x5D4594, 1599608)
	return 0
}

//export sub_502A90
func sub_502A90() *C.char {
	p := (*C.char)(C.dword_5d4594_1599588)
	if C.strlen(p) == 0 {
		return nil
	}
	return p
}

//export sub_502AB0
func sub_502AB0(name *C.char) C.int {
	if name != nil {
		C.strncpy((*C.char)(C.dword_5d4594_1599592), name, 0x7ff)
		return 1
	}
	*(*byte)(C.dword_5d4594_1599592) = *memmap.PtrUint8(0x5D4594, 1599612)
	return 0
}

//export sub_502AF0
func sub_502AF0() *C.char {
	p := (*C.char)(C.dword_5d4594_1599592)
	if C.strlen(p) == 0 {
		return nil
	}
	return p
}

// A FILE* is the existing filesystem registry's native-width handle. Never
// reinterpret it as a Go object or dereference it as a libc FILE structure.
func game4MapReadExact(file *C.FILE, dst []byte) bool {
	for len(dst) != 0 {
		n := int(nox_fs_fread(file, unsafe.Pointer(&dst[0]), int32(len(dst))))
		if n <= 0 || n > len(dst) {
			return false
		}
		dst = dst[n:]
	}
	return true
}

func game4MapReadU32(file *C.FILE) (uint32, bool) {
	var data [4]byte
	if !game4MapReadExact(file, data[:]) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(data[:]), true
}

//export sub_502B10
func sub_502B10() C.int {
	C.dword_5d4594_1599596 = 0
	if C.dword_5d4594_1599588 == nil {
		C.dword_5d4594_1599588 = C.calloc(1, 0x800)
	}
	if C.dword_5d4594_1599592 == nil {
		C.dword_5d4594_1599592 = C.calloc(1, 0x800)
	}
	if C.dword_5d4594_1599576 == nil {
		C.dword_5d4594_1599576 = (*C.char)(C.calloc(1, 0x26000))
	}
	if C.dword_5d4594_1599588 == nil || C.dword_5d4594_1599592 == nil || C.dword_5d4594_1599576 == nil {
		return 0
	}
	first := (*C.char)(C.dword_5d4594_1599588)
	if *(*byte)(unsafe.Pointer(first)) == 0 || C.memchr(unsafe.Pointer(first), 0, 0x800) == nil {
		return 0
	}
	if game4MapOpen(first) == 0 {
		return 0
	}
	defer sub_502DF0()
	if magic, ok := game4MapReadU32(C.nox_file_8); !ok || magic != 0xCAFEDEAD {
		return 0
	}
	for {
		word, ok := game4MapReadU32(C.nox_file_8)
		if !ok {
			return 0
		}
		length := int32(word)
		if length == 0 {
			return 1
		}
		if C.dword_5d4594_1599596 >= 2048 || length < 11 {
			return 0
		}
		afterLength := int64(nox_fs_ftell(C.nox_file_8))
		if afterLength < 4 || uint64(afterLength-4) > math.MaxUint32 {
			return 0
		}
		var one [1]byte
		if !game4MapReadExact(C.nox_file_8, one[:]) || one[0] >= 64 || length < 11+int32(one[0]) {
			return 0
		}
		nameLength := int(one[0])
		var name [64]byte
		if !game4MapReadExact(C.nox_file_8, name[:nameLength]) {
			return 0
		}
		var flags [2]byte
		if !game4MapReadExact(C.nox_file_8, flags[:]) {
			return 0
		}
		x, ok := game4MapReadU32(C.nox_file_8)
		if !ok {
			return 0
		}
		y, ok := game4MapReadU32(C.nox_file_8)
		if !ok {
			return 0
		}
		record := unsafe.Slice((*byte)(game4MapRecord(int32(C.dword_5d4594_1599596))), game4MapRecordBytes)
		// Consume the declared bytes, but emulate strcpy's embedded-NUL stop
		// and leave the remainder of the existing 64-byte name slot intact.
		n := 0
		for name[n] != 0 {
			n++
		}
		copy(record[:], name[:n+1])
		binary.LittleEndian.PutUint32(record[64:], x)
		binary.LittleEndian.PutUint32(record[68:], y)
		binary.LittleEndian.PutUint32(record[72:], uint32(afterLength-4))
		C.dword_5d4594_1599596++
		if nox_fs_fseek(C.nox_file_8, C.long(length-11-int32(nameLength)), int32(C.SEEK_CUR)) != 0 {
			return 0
		}
	}
}

//export sub_502D70
func sub_502D70(index C.int) C.int {
	if index < 0 || int32(index) >= int32(C.dword_5d4594_1599596) {
		return 0
	}
	C.dword_5d4594_3835396 = C.uint32_t(index)
	return C.int(bool2int(nox_xxx_mapgenSaveMap_503830(index) == 0))
}

//export game4MapOpen
func game4MapOpen(name *C.char) C.int {
	file := C.nox_file_8
	if file == nil {
		if nox_xxx_cryptOpen_426910(name, 1, -1) == 0 {
			return 0
		}
		file = nox_xxx_mapgenGetSomeFile_426A60()
		C.nox_file_8 = file
		if file == nil {
			return 0
		}
	}
	nox_fs_fseek(file, 0, int32(C.SEEK_SET))
	return 1
}

//export sub_502DF0
func sub_502DF0() *C.FILE {
	file := C.nox_file_8
	if file != nil {
		nox_xxx_cryptClose_4269F0()
		C.nox_file_8 = nil
	}
	return file
}

//export sub_502E10
func sub_502E10(index C.int) *C.FILE {
	if C.nox_file_8 == nil || index < 0 || int32(index) >= int32(C.dword_5d4594_1599596) {
		return nil
	}
	offset := binary.LittleEndian.Uint32(unsafe.Slice((*byte)(game4MapRecord(int32(index))), game4MapRecordBytes)[72:])
	nox_fs_fseek(C.nox_file_8, C.long(offset), int32(C.SEEK_SET))
	return C.nox_file_8
}

//export sub_502E50
func sub_502E50(name *C.char) *C.FILE {
	if name == nil {
		return nil
	}
	return sub_502E10(sub_5029A0(name))
}

func game4MapCoordinate(index C.int, offset int) C.double {
	if index < 0 || int32(index) >= int32(C.dword_5d4594_1599596) {
		return -1
	}
	bits := binary.LittleEndian.Uint32(unsafe.Slice((*byte)(game4MapRecord(int32(index))), game4MapRecordBytes)[offset:])
	return C.double(math.Float32frombits(bits))
}

//export sub_502E70
func sub_502E70(index C.int) C.double { return game4MapCoordinate(index, 64) }

//export sub_502EA0
func sub_502EA0(index C.int) C.double { return game4MapCoordinate(index, 68) }

//export sub_502ED0
func sub_502ED0(name *C.char) C.int {
	if name == nil || sub_503140() == 0 {
		return 0
	}
	result := nox_mapgenRewriteAreaMap_502ED0(name, (*C.char)(C.dword_5d4594_1599588), (*C.char)(memmap.PtrOff(0x973F18, 42152)))
	if result < 0 {
		return 0
	}
	sub_502B10()
	return result
}

//export sub_503140
func sub_503140() C.int {
	sub_502DF0()
	return nox_mapgenPrepareAreaMap_503140((*C.char)(C.dword_5d4594_1599588), (*C.char)(memmap.PtrOff(0x973F18, 42152)))
}

//export sub_503230
func sub_503230(oldName, newName *C.char) C.int {
	if oldName == nil || newName == nil || sub_503140() == 0 {
		return 0
	}
	if nox_mapgenRenameAreaMap_503230(oldName, newName, (*C.char)(C.dword_5d4594_1599588), (*C.char)(memmap.PtrOff(0x973F18, 42152))) == 0 {
		return 0
	}
	sub_502B10()
	return 1
}

// Read only the common prefix, leaving the file at the attachment/map data.
func game4MapReadExtractPrefix(file *C.FILE) (remaining int32, version byte, ok bool) {
	word, ok := game4MapReadU32(file)
	if !ok {
		return 0, 0, false
	}
	length := int32(word)
	var one [1]byte
	if !game4MapReadExact(file, one[:]) || one[0] >= 64 || length < 11+int32(one[0]) {
		return 0, 0, false
	}
	var name [64]byte
	if !game4MapReadExact(file, name[:one[0]]) {
		return 0, 0, false
	}
	var flags [2]byte
	var coordinates [8]byte
	if !game4MapReadExact(file, flags[:]) || !game4MapReadExact(file, coordinates[:]) {
		return 0, 0, false
	}
	return length - 11 - int32(one[0]), flags[1], true
}

func game4MapCopyExact(input, output *C.FILE, count int32) bool {
	var buffer [4096]byte
	for count > 0 {
		n := count
		if n > int32(len(buffer)) {
			n = int32(len(buffer))
		}
		if !game4MapReadExact(input, buffer[:n]) || nox_fs_fwrite(output, unsafe.Pointer(&buffer[0]), n) != n {
			return false
		}
		count -= n
	}
	return true
}

//export sub_5034B0
func sub_5034B0(name *C.char) C.int {
	source := sub_502A90()
	if name == nil || source == nil {
		return 0
	}
	directory := unsafe.Slice((*byte)(memmap.PtrOff(0x973F18, 42152)), 0x800)
	dirLength := 0
	for dirLength < len(directory) && directory[dirLength] != 0 {
		dirLength++
	}
	const suffix = "\\copy.tmp\x00"
	if dirLength == 0 || dirLength+len(suffix) > 0x800 {
		return 0
	}
	var path [0x800]byte
	copy(path[:], directory[:dirLength])
	copy(path[dirLength:], suffix)
	outputPath := (*C.char)(unsafe.Pointer(&path[0]))
	game4MapOpen(source)
	input := sub_502E50(name)
	if input == nil {
		sub_502DF0()
		return 0
	}
	remaining, version, ok := game4MapReadExtractPrefix(input)
	if !ok {
		sub_502DF0()
		return 0
	}
	if version > 1 {
		if remaining < 4 {
			sub_502DF0()
			return 0
		}
		extra, read := game4MapReadU32(input)
		remaining -= 4
		if !read || int32(extra) < 0 || int32(extra) > remaining || nox_fs_fseek(input, C.long(extra), int32(C.SEEK_CUR)) != 0 {
			sub_502DF0()
			return 0
		}
		remaining -= int32(extra)
	}
	if remaining < 4 {
		sub_502DF0()
		return 0
	}
	magic, ok := game4MapReadU32(input)
	if !ok || magic != 0xCAFEDEAD {
		sub_502DF0()
		return 0
	}
	remaining -= 4
	output := nox_fs_create(outputPath)
	if output == nil {
		sub_502DF0()
		return 0
	}
	header := [4]byte{0xCE, 0xFA, 0xED, 0xAB}
	ok = nox_fs_fwrite(output, unsafe.Pointer(&header[0]), 4) == 4 && game4MapCopyExact(input, output, remaining)
	sub_502DF0()
	nox_fs_close(output)
	if !ok {
		nox_fs_remove(outputPath)
	}
	return C.int(bool2int(ok))
}

//export sub_5036D0
func sub_5036D0(name, outputPath *C.char) C.int {
	if name == nil || outputPath == nil {
		return 0
	}
	// This removal precedes source lookup, even on failure.
	nox_fs_remove(outputPath)
	source := sub_502A90()
	if source == nil {
		return 0
	}
	game4MapOpen(source)
	input := sub_502E50(name)
	if input == nil {
		sub_502DF0()
		return 0
	}
	remaining, version, ok := game4MapReadExtractPrefix(input)
	if !ok || version <= 1 || remaining < 4 {
		sub_502DF0()
		return 0
	}
	extra, ok := game4MapReadU32(input)
	if !ok || int32(extra) <= 0 || int32(extra) > remaining-4 {
		sub_502DF0()
		return 0
	}
	output := nox_fs_create(outputPath)
	if output == nil {
		sub_502DF0()
		return 0
	}
	ok = game4MapCopyExact(input, output, int32(extra))
	sub_502DF0()
	nox_fs_close(output)
	if !ok {
		nox_fs_remove(outputPath)
	}
	return C.int(bool2int(ok))
}

//export nox_xxx_mapgenSaveMap_503830
func nox_xxx_mapgenSaveMap_503830(index C.int) C.int {
	if index < 0 || int32(index) >= int32(C.dword_5d4594_1599596) {
		return 0
	}
	C.nox_xxx_free_503F40()
	*memmap.PtrUint32(0x5D4594, 1599572) = math.MaxUint32
	C.dword_5d4594_1599644 = 0
	if game4MapOpen((*C.char)(C.dword_5d4594_1599588)) == 0 || sub_502E10(index) == nil || !mapgenLoadCurrent503830() {
		sub_502DF0()
		return 0
	}
	C.dword_5d4594_1599480 = C.uint32_t(index)
	C.dword_5d4594_1599476 = 0
	C.dword_5d4594_3835396 = C.uint32_t(index)
	sub_502DF0()
	return 1
}

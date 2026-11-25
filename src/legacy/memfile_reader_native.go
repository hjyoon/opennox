package legacy

/*
#include "memfile.h"
*/
import "C"

import (
	"encoding/binary"
	"io"
	"unsafe"
)

func memfileRemaining(file *C.nox_memfile) uintptr {
	if file == nil || file.data == nil || file.cur == nil {
		return 0
	}
	start := uintptr(unsafe.Pointer(file.data))
	current, end := uintptr(unsafe.Pointer(file.cur)), uintptr(unsafe.Pointer(file.end))
	if current < start || current > end {
		return 0
	}
	return end - current
}

func memfileTake(file *C.nox_memfile, size uintptr) []byte {
	if memfileRemaining(file) < size {
		return nil
	}
	start := unsafe.Pointer(file.cur)
	file.cur = (*C.char)(unsafe.Add(start, size))
	return unsafe.Slice((*byte)(start), int(size))
}

//export nox_memfile_read_u8
func nox_memfile_read_u8(file *C.nox_memfile) C.uint8_t {
	data := memfileTake(file, 1)
	if data == nil {
		return 0
	}
	return C.uint8_t(data[0])
}

//export nox_memfile_read_i8
func nox_memfile_read_i8(file *C.nox_memfile) C.int8_t {
	return C.int8_t(nox_memfile_read_u8(file))
}

//export nox_memfile_read_u16
func nox_memfile_read_u16(file *C.nox_memfile) C.uint16_t {
	data := memfileTake(file, 2)
	if data == nil {
		return 0
	}
	return C.uint16_t(binary.LittleEndian.Uint16(data))
}

//export nox_memfile_read_i16
func nox_memfile_read_i16(file *C.nox_memfile) C.int16_t {
	return C.int16_t(nox_memfile_read_u16(file))
}

//export nox_memfile_read_u32
func nox_memfile_read_u32(file *C.nox_memfile) C.uint32_t {
	data := memfileTake(file, 4)
	if data == nil {
		return 0
	}
	return C.uint32_t(binary.LittleEndian.Uint32(data))
}

//export nox_memfile_read_i32
func nox_memfile_read_i32(file *C.nox_memfile) C.int32_t {
	return C.int32_t(nox_memfile_read_u32(file))
}

//export nox_memfile_skip
func nox_memfile_skip(file *C.nox_memfile, count C.int) {
	if file != nil && file.data != nil && file.cur != nil {
		_, _ = asMemfile(file).Seek(int64(count), io.SeekCurrent)
	}
}

//export nox_memfile_read
func nox_memfile_read(destination unsafe.Pointer, size C.uint, count C.int, file *C.nox_memfile) C.uint {
	n := min(uintptr(uint32(size)*uint32(count)), memfileRemaining(file))
	if n == 0 {
		return 0
	}
	source := unsafe.Pointer(file.cur)
	copy(unsafe.Slice((*byte)(destination), int(n)), unsafe.Slice((*byte)(source), int(n)))
	file.cur = (*C.char)(unsafe.Add(source, n))
	return C.uint(n / uintptr(size))
}

func memfileReadAligned(destination []byte, file *C.nox_memfile) C.uint {
	if file == nil || file.data == nil {
		return 0
	}
	var data [8]byte
	offset := (uintptr(unsafe.Pointer(file.cur)) - uintptr(unsafe.Pointer(file.data))) % 8
	if offset != 0 {
		nox_memfile_skip(file, C.int(8-offset))
	}
	if nox_memfile_read(unsafe.Pointer(&data[0]), 8, 1, file) != 1 {
		return 0
	}
	copy(destination, data[:])
	return 1
}

//export nox_memfile_read64align_40AD60
func nox_memfile_read64align_40AD60(destination *C.char, size, count C.int, file *C.nox_memfile) C.uint {
	length := int64(size) * int64(count)
	if length < 0 || length > 8 {
		return 0
	}
	return memfileReadAligned(unsafe.Slice((*byte)(unsafe.Pointer(destination)), int(length)), file)
}

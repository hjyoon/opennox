package legacy

/*
#include "GAME2_2.h"

*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_476080
func sub_476080(record *C.uchar) C.int {
	wall := unsafe.Slice(record, 7)
	y := 23 * int32(wall[6])
	player := asDrawable((*nox_drawable)(*memmap.PtrPtr(0x852978, 8)))
	if player == nil {
		return C.int(y + 11)
	}
	x := 23 * int32(wall[5])
	dx, dy := int32(player.PosVec.X)-x, int32(player.PosVec.Y)-y
	var side int32
	switch wall[0] {
	case 0, 3, 11:
		side = 23 * (dx - 22 + dy)
	case 1, 4, 12:
		side = 23 * (dy - dx)
	default:
		return C.int(y + 11)
	}
	if side < 0 {
		y += 22
	}
	return C.int(y)
}

//export sub_47D380
func sub_47D380(a, b C.int) C.int {
	clip := currentDrawData.clip
	minimum, maximum := C.int(clip.min_x), C.int(clip.max_x)
	left, right := min(a, b), max(a, b)
	if left >= maximum || right < minimum {
		return 0
	}
	left, right = max(left, minimum), min(right, maximum)
	if left == right {
		return 0
	}
	if left != minimum || right != maximum {
		*memmap.PtrInt32(0x973F18, 52) = int32(left)
		*memmap.PtrInt32(0x973F18, 12) = int32(right)
		Set_dword_5d4594_3799452(1)
	}
	return 1
}

//export sub_47DBC0
func sub_47DBC0() C.uchar { return C.uchar(memmap.Uint8(0x5D4594, 1193128)) }

//export sub_47FCE0
func sub_47FCE0(descriptor *C.uint32_t, unused C.int) C.int {
	fields := unsafe.Slice(descriptor, 27)
	key := [4]uint32{uint32(fields[3]), uint32(fields[2]), uint32(fields[21]), uint32(fields[26])}
	count := int32(legacyAddressableWords.dword_5d4594_3804684)
	for i := range count {
		if *memmap.PtrT[[4]uint32](0x973F18, uintptr(6088+16*i)) == key {
			return 1
		}
	}
	*memmap.PtrT[[4]uint32](0x973F18, uintptr(6088+16*count)) = key
	legacyAddressableWords.dword_5d4594_3804684++
	return 1
}

//export sub_480220
func sub_480220(rgb *C.uint8_t, packed *C.uint8_t) *C.uint8_t {
	out := unsafe.Slice((*byte)(unsafe.Pointer(rgb)), 3)
	in := unsafe.Slice((*byte)(unsafe.Pointer(packed)), 2)
	out[0] = in[0] << 3
	out[1] = byte(binary.LittleEndian.Uint16(in)>>3) & 0xfc
	out[2] = in[1] & 0xf8
	return packed
}

//export sub_480250
func sub_480250(rgb *C.uint8_t, packed *C.uint16_t) *C.uint16_t {
	in := unsafe.Slice((*byte)(unsafe.Pointer(rgb)), 3)
	pixel := uint16(in[0]>>3) | uint16(in[1]&0xfc)<<3 | uint16(in[2]&0xf8)<<8
	binary.LittleEndian.PutUint16(unsafe.Slice((*byte)(unsafe.Pointer(packed)), 2), pixel)
	return packed
}

//export nox_xxx_setSomeFunc_48A210
func nox_xxx_setSomeFunc_48A210(callback unsafe.Pointer) unsafe.Pointer {
	*memmap.PtrPtr(0x5D4594, 1193504) = callback
	return callback
}

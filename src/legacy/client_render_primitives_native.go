package legacy

/*
#include "GAME2_3.h"
#include "common/alloc/classes/alloc_class.h"
#include <stdlib.h>

extern void* nox_alloc_pixelSpan_1301844;
extern void* dword_5d4594_1301848;
*/
import "C"

import (
	"encoding/binary"
	"image"
	"unsafe"
)

//export sub_49AEA0
func sub_49AEA0() C.int {
	if pool := C.nox_alloc_pixelSpan_1301844; pool != nil {
		nox_free_alloc_class((*C.nox_alloc_class)(pool))
		C.nox_alloc_pixelSpan_1301844 = nil
	}
	if pixels := C.dword_5d4594_1301848; pixels != nil {
		C.free(pixels)
		C.dword_5d4594_1301848 = nil
	}
	return 1
}

//export sub_49CD30
func sub_49CD30(x, y, width, height, color, level C.int) {
	if width == 0 || height == 0 {
		return
	}
	r := GetClient().R2()
	draw := r.Data()
	right, bottom := x+width, y+height
	if draw.Clip() {
		clip := draw.ClipRect()
		if max(int(x), clip.Min.X) >= min(int(right), clip.Max.X) ||
			max(int(y), clip.Min.Y) >= min(int(bottom), clip.Max.Y) {
			return
		}
	}
	nox_draw_set54RGB32_434040(int32(color))
	sub_434080(int32(level))
	line := func(x0, y0, x1, y1 C.int) {
		r.AddPoint(image.Pt(int(x0), int(y0)))
		r.AddPoint(image.Pt(int(x1), int(y1)))
		r.DrawParticles49ED80(64)
	}
	line(x, y, right-1, y)
	line(right, y, right, bottom-1)
	line(right, bottom, x, bottom)
	line(x, bottom-1, x, y+1)
}

//export sub_49D1C0
func sub_49D1C0(dst unsafe.Pointer, color, size C.int) C.int {
	fillPixelWords(dst, uint32(color), int32(size))
	return 0
}

//export sub_49E3C0
func sub_49E3C0(dst *C.uint32_t, color C.int, size C.uint) {
	fillPixelWords(unsafe.Pointer(dst), uint32(color), int32(size))
}

func fillPixelWords(dst unsafe.Pointer, color uint32, size int32) {
	if size <= 0 {
		return
	}
	pixels := unsafe.Slice((*byte)(dst), int(size))
	for len(pixels) >= 4 {
		binary.LittleEndian.PutUint32(pixels, color)
		pixels = pixels[4:]
	}
	if len(pixels) >= 2 {
		binary.LittleEndian.PutUint16(pixels, uint16(color))
	}
}

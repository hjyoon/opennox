package legacy

/*
#include "GAME1_2.h"
#include "GAME2_3.h"
*/
import "C"

import (
	"cmp"
	"encoding/binary"
	"slices"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func sortPaletteColors(colors [][4]byte) {
	// Pack the original index into the low byte and sort by the unsigned
	// little-endian word, matching the order used by the display palette.
	slices.SortFunc(colors, func(a, b [4]byte) int {
		return cmp.Compare(binary.LittleEndian.Uint32(a[:]), binary.LittleEndian.Uint32(b[:]))
	})
}

//export sub_48C580
func sub_48C580(colors *C.pixel8888, count C.int) {
	if count > 0 {
		sortPaletteColors(unsafe.Slice((*[4]byte)(unsafe.Pointer(colors)), int(count)))
	}
}

//export sub_435040
func sub_435040() {
	source := memmap.PtrT[[256][4]byte](0x973F18, 3880)
	var sorted [256][4]byte
	for i, color := range source {
		sorted[i] = [4]byte{byte(i), color[0], color[1], color[2]}
	}
	sortPaletteColors(sorted[:])
	palette := memmap.PtrT[[256][4]byte](0x5D4594, 809604)
	order := memmap.PtrT[[256]byte](0x5D4594, 808304)
	for i, color := range sorted {
		palette[i] = [4]byte{color[1], color[2], color[3], 0}
		order[i] = color[0]
	}
}

//export sub_435120
func sub_435120(dst, src unsafe.Pointer) {
	out, in := (*[256][4]byte)(dst), (*[256][4]byte)(src)
	for i := range out {
		out[i][0] = in[i][0]
		out[i][1] = in[i][1]
		out[i][2], out[i][3] = in[i][2], 4
	}
}

//export sub_435150
func sub_435150(dst *C.uint8_t, src *C.char) {
	out, in := (*[256][3]byte)(unsafe.Pointer(dst)), (*[256][4]byte)(unsafe.Pointer(src))
	for i := range out {
		out[i][0] = in[i][0]
		out[i][1] = in[i][1]
		out[i][2] = in[i][2]
	}
}

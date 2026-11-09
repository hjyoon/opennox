package legacy

/*
#include "defs.h"
#include "memfile.h"
#include "GAME1.h"
#include "GAME2.h"
extern uint32_t nox_tile_def_cnt;
extern nox_tileDef_t nox_tile_defs_arr[176];
extern uint32_t dword_5d4594_251572;
*/
import "C"

import (
	"encoding/binary"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

var _ = [1]struct{}{}[unsafe.Sizeof(server.TileDef{})-C.sizeof_nox_tileDef_t]
var _ = [1]struct{}{}[C.sizeof_nox_tileDef_t-unsafe.Sizeof(server.TileDef{})]

func thingTileCountCell411540() *uint32 { return (*uint32)(unsafe.Pointer(&C.nox_tile_def_cnt)) }
func thingEdgeCountCell411850() *uint32 { return (*uint32)(unsafe.Pointer(&C.dword_5d4594_251572)) }
func thingTileOwner411540(i uint32) *server.TileDef {
	return (*server.TileDef)(unsafe.Pointer(&C.nox_tile_defs_arr[i]))
}
func thingEdgeImageCell411850(i uint32) *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(&C.nox_edge_images_native[i]))
}
func thingEdgeBytes411850(i uint32, off uintptr, n int) []byte {
	return unsafe.Slice((*byte)(memmap.PtrOff(0x85B3FC, off+60*uintptr(i))), n)
}
func thingEdgeStore32411850(i uint32, off uintptr, v uint32) {
	binary.LittleEndian.PutUint32(thingEdgeBytes411850(i, off, 4), v)
}
func thingEdgeStore16411850(i uint32, off uintptr, v uint16) {
	binary.LittleEndian.PutUint16(thingEdgeBytes411850(i, off, 2), v)
}
func thingFloorScratchByte411540() byte { return *memmap.PtrUint8(0x5D4594, 251576) }
func thingEdgeScratchByte411850() byte  { return *memmap.PtrUint8(0x5D4594, 251580) }
func thingParserColor411540(r, g, b byte) uint32 {
	return uint32(nox_color_rgb_4344A0(int32(r), int32(g), int32(b)))
}

type thingParserServices411540 struct {
	Facade func(*server.TileDef) int32
	Audio  func(*nox_memfile, unsafe.Pointer) int32
}

var thingParserServicesCurrent411540 *thingParserServices411540

func pushThingParserServices411540(o *thingParserServices411540) func() {
	previous := thingParserServicesCurrent411540
	thingParserServicesCurrent411540 = o
	return func() {
		if thingParserServicesCurrent411540 != o {
			panic("thing parser observer pop order")
		}
		thingParserServicesCurrent411540 = previous
	}
}
func thingParserFacade411540(p *server.TileDef) int32 {
	if o := thingParserServicesCurrent411540; o != nil && o.Facade != nil {
		return o.Facade(p)
	}
	return int32(C.nox_xxx_checkFacade_4117E0((*C.nox_tileDef_t)(unsafe.Pointer(p))))
}
func thingParserAudio415660(f *nox_memfile, scratch unsafe.Pointer) int32 {
	if o := thingParserServicesCurrent411540; o != nil && o.Audio != nil {
		return o.Audio(f, scratch)
	}
	return int32(C.sub_452BD0(f, (*C.char)(scratch)))
}

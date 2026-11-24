package legacy

/*
#include "GAME2_2.h"
#include "floor_tile_storage.h"
#include "client__draw__staticdraw.h"
#include "client__draw__parse__parse.h"

extern void* nox_video_tileBuf_ptr_3798796;
extern void* nox_video_tileBuf_end_3798844;
extern uint32_t dword_5d4594_3798800, dword_5d4594_3798804, dword_5d4594_3798808;
extern uint32_t dword_5d4594_3798820, dword_5d4594_3798824, dword_5d4594_3798836, dword_5d4594_3798840;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

const floorTileRows = 46

func floorTileBuffer() []byte {
	begin, end := C.nox_video_tileBuf_ptr_3798796, C.nox_video_tileBuf_end_3798844
	if begin == nil || uintptr(end) <= uintptr(begin) {
		return nil
	}
	return unsafe.Slice((*byte)(begin), int(uintptr(end)-uintptr(begin)))
}

func floorBufferIndex(offset, size int) int {
	index := offset % size
	if index < 0 {
		index += size
	}
	return index
}

func copyFloorSpan(buffer []byte, offset int, source []byte) {
	index := floorBufferIndex(offset, len(buffer))
	for len(source) != 0 {
		count := copy(buffer[index:], source)
		source = source[count:]
		index = 0
	}
}

func fillFloorSpan(buffer []byte, offset int, color uint32, size int) {
	index := floorBufferIndex(offset, len(buffer))
	var pattern [4]byte
	binary.LittleEndian.PutUint32(pattern[:], color)
	for i := range size {
		buffer[index] = pattern[i&3]
		index++
		if index == len(buffer) {
			index = 0
		}
	}
}

func floorTileBase(x, y int32, shift uint8) int {
	dx := int32(C.dword_5d4594_3798836) + x - int32(C.dword_5d4594_3798820)
	dy := int32(C.dword_5d4594_3798840) + y - int32(C.dword_5d4594_3798824)
	return int(C.dword_5d4594_3798804)*int(dy) + int(dx<<shift)
}

func floorTileRow(base, row int, shift uint8) (offset, size int) {
	offset = base + row*int(C.dword_5d4594_3798804) + int(memmap.Uint32(0x973CE0, 192+4*uintptr(row)))<<shift
	size = int(memmap.Uint32(0x973CE0, 384+4*uintptr(row))) << shift
	return offset, size
}

//export nox_xxx_tileDraw_4815E0
func nox_xxx_tileDraw_4815E0(pos *C.int2, image *nox_video_bag_image_t, tileIndex C.uint16_t) {
	buffer := floorTileBuffer()
	pixels := asImage(image).Pixdata()
	if len(buffer) == 0 || len(pixels) == 0 {
		return
	}
	shift := memmap.Uint8(0x973F18, 7696)
	base := floorTileBase(int32(pos.field_0), int32(pos.field_4), shift)
	for row := range floorTileRows {
		offset, size := floorTileRow(base, row, shift)
		if size < 0 || size > len(pixels) {
			return
		}
		copyFloorSpan(buffer, offset, pixels[:size])
		pixels = pixels[size:]
	}
}

//export sub_481770
func sub_481770(pos *C.int2, image *nox_video_bag_image_t, tileIndex C.uint16_t) {
	tiles := nox_floor_tile_storage()
	buffer := floorTileBuffer()
	if C.uint32_t(tileIndex) >= tiles.count || len(buffer) == 0 {
		return
	}
	tile := &tiles.definitions[tileIndex]
	if !bool(get_nox_client_texturedFloors_154956()) && tile.field_58&1 != 0 {
		Set_dword_5d4594_1193156(1)
	}
	shift := memmap.Uint8(0x973F18, 7696)
	base := floorTileBase(int32(pos.field_0), int32(pos.field_4), shift)
	for row := range floorTileRows {
		offset, size := floorTileRow(base, row, shift)
		fillFloorSpan(buffer, offset, uint32(tile.color_48), size)
	}
}

//export nox_xxx_drawNoEdges_4818F0
func nox_xxx_drawNoEdges_4818F0(pos *C.int2, edge *C.nox_tile_list_node_t) {}

func floorDrawImageAt(images unsafe.Pointer, index int) *nox_video_bag_image_t {
	return *(**nox_video_bag_image_t)(unsafe.Add(images, uintptr(index)*unsafe.Sizeof(uintptr(0))))
}

//export nox_xxx_drawTexEdgesProbably_481900
func nox_xxx_drawTexEdgesProbably_481900(pos *C.int2, edge *C.nox_tile_list_node_t) {
	tiles := nox_floor_tile_storage()
	buffer := floorTileBuffer()
	if pos == nil || edge == nil || edge.field_0 < 0 || C.uint32_t(edge.field_0) >= tiles.count ||
		edge.field_2 < 0 || edge.field_2 >= 64 || len(buffer) == 0 {
		return
	}
	tile := &tiles.definitions[edge.field_0]
	images := C.nox_edge_images_native[edge.field_2]
	if tile.data_32 == nil || images == nil {
		return
	}
	tileIndex := int(edge.field_1) + int(tile.field_46)
	edgeIndex := int(edge.field_3) + int(memmap.Uint16(0x85B3FC, 28690+60*uintptr(edge.field_2)))
	if tileIndex < 0 || edgeIndex < 0 {
		return
	}
	commands := asImage(floorDrawImageAt(images, edgeIndex)).Pixdata()
	texture := asImage(floorDrawImageAt(unsafe.Pointer(tile.data_32), tileIndex)).Pixdata()
	if len(commands) < 2 || len(texture) == 0 {
		return
	}
	first, last := int(commands[0]), int(commands[1])
	if first > last || last >= floorTileRows {
		return
	}
	commands = commands[2:]
	*memmap.PtrUint32(0x5D4594, 2523980+4*uintptr(edge.field_2)) = 1
	shift := memmap.Uint8(0x973F18, 7696)
	base := floorTileBase(int32(pos.field_0), int32(pos.field_4), shift)
	for row := first; row <= last; row++ {
		dst, _ := floorTileRow(base, row, shift)
		src := int(memmap.Uint32(0x973CE0, 4*uintptr(row))) << shift
		remaining := int(int32(memmap.Uint32(0x973CE0, 384+4*uintptr(row))))
		for remaining > 0 {
			if len(commands) < 2 {
				return
			}
			kind, count := commands[0], int(commands[1])
			commands = commands[2:]
			if count == 0 || count > remaining {
				return
			}
			size := count << shift
			switch kind {
			case 1:
			case 2:
				if size < 0 || size > len(commands) {
					return
				}
				copyFloorSpan(buffer, dst, commands[:size])
				commands = commands[size:]
			case 3:
				if src < 0 || src > len(texture) || size < 0 || size > len(texture)-src {
					return
				}
				copyFloorSpan(buffer, dst, texture[src:src+size])
			default:
				return
			}
			dst, src, remaining = dst+size, src+size, remaining-count
		}
	}
}

//export sub_476AE0
func sub_476AE0(view *nox_draw_viewport_t, dr *nox_drawable) {
	buffer := floorTileBuffer()
	if dr == nil || dr.field_76 == nil || len(buffer) == 0 {
		return
	}
	var image *nox_video_bag_image_t
	if unsafe.Pointer(dr.draw_func) == C.nox_thing_static_draw {
		if dr.flags28&0x40000 != 0 && dr.flags30&0x1000000 == 0 {
			return
		}
		image = (*C.nox_static_draw_data_t)(dr.field_76).image
	} else {
		data := (*C.nox_static_random_draw_data_t)(dr.field_76)
		if data.images == nil || dr.field_77 >= C.uint32_t(data.count) {
			return
		}
		image = floorDrawImageAt(unsafe.Pointer(data.images), int(dr.field_77))
	}
	pixels := asImage(image).Pixdata()
	if len(pixels) < 17 {
		return
	}
	read := binary.LittleEndian.Uint32
	width, height := int32(read(pixels)), int32(read(pixels[4:]))
	x := int32(read(pixels[8:])) + int32(dr.pos.x) - int32(byte(dr.field_0))
	y := int32(read(pixels[12:])) + int32(dr.pos.y) - int32(dr.field_26_1) - int32(dr.z) - int32(byte(dr.field_0>>8))
	originX, originY := int32(C.dword_5d4594_3798820), int32(C.dword_5d4594_3798824)
	if x < originX || x+width >= originX+int32(C.dword_5d4594_3798800) ||
		y < originY || y+height >= originY+int32(C.dword_5d4594_3798808) {
		dr.field_86 = 0
		return
	}
	counter := int32(Get_nox_xxx_waypointCounterMB_587000_154948())
	if counter <= 0 {
		dr.field_86 = 0
	} else if counter-int32(dr.field_86) <= 1 {
		dr.field_86 = C.uint32_t(counter)
		return
	}

	runs := pixels[17:]
	rowOffset := floorTileBase(x, y, 1)
	for range height {
		dst, remaining := rowOffset, width
		for remaining > 0 {
			if len(runs) < 2 {
				return
			}
			kind, count := runs[0]&0xf, int32(runs[1])
			runs = runs[2:]
			if count == 0 || count > remaining {
				return
			}
			size := 2 * int(count)
			switch kind {
			case 1:
				dst += size
			case 2:
				if size > len(runs) {
					return
				}
				copyFloorSpan(buffer, dst, runs[:size])
				runs = runs[size:]
				dst += size
			}
			remaining -= count
		}
		rowOffset += int(C.dword_5d4594_3798804)
	}
}

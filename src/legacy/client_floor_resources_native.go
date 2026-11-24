package legacy

/*
#include <stdlib.h>
#include "GAME2_2.h"
#include "floor_tile_storage.h"

extern uint32_t dword_5d4594_3798804;
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func readFloorImageRefs(f *binfile.MemFile, name []byte, count int, imageType, defaultName byte) (**nox_video_bag_image_t, bool) {
	if count > len(f.Data())/4 {
		return nil, false
	}
	images := (**nox_video_bag_image_t)(C.calloc(C.size_t(count), C.size_t(unsafe.Sizeof(uintptr(0)))))
	if images == nil && count != 0 {
		return nil, false
	}
	references := unsafe.Slice(images, count)
	for i := range references {
		if len(f.Data()) < 4 {
			C.free(unsafe.Pointer(images))
			return nil, false
		}
		id := f.ReadI32()
		name[0] = defaultName
		if id == -1 {
			if len(f.Data()) < 2 {
				C.free(unsafe.Pointer(images))
				return nil, false
			}
			imageType = f.ReadU8()
			length := int(f.ReadU8())
			if length >= len(name) || length > len(f.Data()) {
				C.free(unsafe.Pointer(images))
				return nil, false
			}
			copy(name, f.Data()[:length])
			name[length] = 0
			f.Skip(length)
		}
		references[i] = nox_xxx_readImgMB_42FAA0(C.int(id), C.char(imageType), (*C.char)(unsafe.Pointer(&name[0])))
	}
	return images, true
}

func readFloorTileImages(f *binfile.MemFile, scratch []byte) int {
	if len(f.Data()) < 5 {
		return 0
	}
	f.Skip(4)
	rawName, err := f.ReadBytes8()
	if err != nil {
		return 0
	}
	name := alloc.GoStringS(rawName)
	tiles := nox_floor_tile_storage()
	index := len(rawName)
	if tiles.count > 0 {
		index = -1
		for i := range min(int(tiles.count), len(tiles.definitions)) {
			tileName := unsafe.Slice((*byte)(unsafe.Pointer(&tiles.definitions[i].name[0])), len(tiles.definitions[i].name))
			if alloc.GoStringS(tileName) == name {
				index = i
				break
			}
		}
	}
	_, debug := os.LookupEnv("NOX_DEBUG_THINGS")
	if debug {
		fmt.Fprintf(os.Stderr, "[things-floor] name=%s len=%d\n", name, len(rawName))
	}
	if index < 0 || index >= len(tiles.definitions) || len(f.Data()) < 16 {
		return 0
	}
	f.Skip(12)
	imageType, width, height := f.ReadU8(), f.ReadU8(), f.ReadU8()
	f.Skip(1)
	count := int(imageType) * int(width) * int(height)
	if debug {
		fmt.Fprintf(os.Stderr, "[things-floor] index=%d type=%d width=%d height=%d count=%d\n", index, imageType, width, height, count)
	}
	images, ok := readFloorImageRefs(f, scratch, count, imageType, memmap.Uint8(0x5D4594, 1193192))
	if !ok {
		return 0
	}
	tiles.definitions[index].data_32 = images
	if debug {
		fmt.Fprintf(os.Stderr, "[things-floor] images=%p\n[things-floor] done index=%d\n", images, index)
	}
	return 1
}

func readFloorEdgeImages(f *binfile.MemFile, scratch []byte) int {
	if len(f.Data()) < 5 {
		return 0
	}
	f.Skip(4)
	rawName, err := f.ReadBytes8()
	if err != nil {
		return 0
	}
	name := alloc.GoStringS(rawName)
	count := min(int(*thingEdgeCountCell411850()), len(C.nox_edge_images_native))
	index := 0
	for index < count && GoStringP(memmap.PtrOff(0x85B3FC, 28644+60*uintptr(index))) != name {
		index++
	}
	if index == count || len(f.Data()) < 14 {
		return 0
	}
	f.Skip(9)
	imageType := f.ReadU8()
	f.Skip(1)
	if f.ReadU8() == 1 {
		return 0
	}
	imageCount := 2 * int(imageType) * (int(f.ReadU8()) + int(f.ReadU8()))
	images, ok := readFloorImageRefs(f, scratch, imageCount, imageType, memmap.Uint8(0x5D4594, 1193196))
	if !ok || images == nil {
		return 0
	}
	if len(f.Data()) < 4 || f.ReadU32() != 0x454e4420 {
		C.free(unsafe.Pointer(images))
		return 0
	}
	C.nox_edge_images_native[index] = unsafe.Pointer(images)
	*memmap.PtrUint32(0x85B3FC, 28676+60*uintptr(index)) = 0
	return 1
}

//export nox_thing_read_floor_485B30
func nox_thing_read_floor_485B30(file *nox_memfile, scratch *C.char) C.int {
	return C.int(readFloorTileImages((*binfile.MemFile)(unsafe.Pointer(file)), unsafe.Slice((*byte)(unsafe.Pointer(scratch)), 256)))
}

//export nox_thing_read_edge_485D40
func nox_thing_read_edge_485D40(file *nox_memfile, scratch *C.char) C.int {
	return C.int(readFloorEdgeImages((*binfile.MemFile)(unsafe.Pointer(file)), unsafe.Slice((*byte)(unsafe.Pointer(scratch)), 256)))
}

//export nox_xxx_tile_486060
func nox_xxx_tile_486060() C.int {
	Set_dword_5d4594_1193188(1)
	stride := 45 * uint32(C.dword_5d4594_3798804)
	*memmap.PtrUint32(0x973CE0, 376) = stride + uint32(46)<<memmap.Uint8(0x973F18, 7696)
	return C.int(stride)
}

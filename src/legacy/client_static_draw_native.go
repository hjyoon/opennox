package legacy

/*
#include "cgo_const.h"
#include "client__draw__parse__parse.h"
#include "client__draw__staticdraw.h"
#include "draw_allocation.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

func readLegacyDrawImage(file *C.nox_memfile, scratch *C.char, nameOffset uintptr) *nox_video_bag_image_t {
	imageID := C.int(nox_memfile_read_i32(file))
	*scratch = C.char(memmap.Uint8(0x5D4594, nameOffset))
	var kind C.char
	if imageID == -1 {
		kind = C.char(nox_memfile_read_u8(file))
		length := nox_memfile_read_u8(file)
		nox_memfile_read(unsafe.Pointer(scratch), 1, C.int(length), file)
		*(*C.char)(unsafe.Add(unsafe.Pointer(scratch), uintptr(length))) = 0
	}
	return nox_xxx_readImgMB_42FAA0(imageID, kind, scratch)
}

func legacyDrawImageList(file *C.nox_memfile, scratch *C.char, count int, nameOffset uintptr) (**nox_video_bag_image_t, bool) {
	if count == 0 {
		return nil, true
	}
	base := (**nox_video_bag_image_t)(C.nox_draw_memory_calloc(C.size_t(count), C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)))))
	if base == nil {
		return nil, false
	}
	images := unsafe.Slice(base, count)
	for i := range images {
		images[i] = readLegacyDrawImage(file, scratch, nameOffset)
	}
	return base, true
}

//export nox_xxx_spriteLoadStaticRandomData_44C000
func nox_xxx_spriteLoadStaticRandomData_44C000(scratch *C.char, file *C.nox_memfile) *C.nox_static_random_draw_data_t {
	size := unsafe.Sizeof(C.nox_static_random_draw_data_t{})
	data := (*C.nox_static_random_draw_data_t)(C.nox_draw_memory_calloc(1, C.size_t(size)))
	if data == nil {
		return nil
	}
	data.size = C.uint32_t(size)
	data.count = nox_memfile_read_u8(file)
	images, ok := legacyDrawImageList(file, scratch, int(data.count), 830852)
	if !ok {
		C.nox_draw_memory_free(unsafe.Pointer(data))
		return nil
	}
	data.images = images
	return data
}

//export get_animation_kind_id_44B4C0
func get_animation_kind_id_44B4C0(name *C.nox_cgo_const_char) C.int {
	return C.int(client.ParseAnimKind(GoString((*C.char)(unsafe.Pointer(name)))))
}

//export nox_thing_static_draw
func nox_thing_static_draw(viewport *C.uint32_t, drawable *C.nox_drawable) C.int {
	data := (*C.nox_static_draw_data_t)(drawable.field_76)
	if data != nil && (drawable.flags28&0x40000 == 0 || drawable.flags30&0x1000000 != 0) {
		nox_xxx_drawObject_4C4770_draw((*C.nox_draw_viewport_t)(unsafe.Pointer(viewport)), drawable, unsafe.Pointer(data.image))
	}
	return 1
}

//export nox_thing_static_random_draw
func nox_thing_static_random_draw(viewport *C.uint32_t, drawable *C.nox_drawable) C.int {
	data := (*C.nox_static_random_draw_data_t)(drawable.field_76)
	if data != nil && data.images != nil && drawable.field_77 < C.uint32_t(data.count) {
		image := unsafe.Slice(data.images, int(data.count))[drawable.field_77]
		nox_xxx_drawObject_4C4770_draw((*C.nox_draw_viewport_t)(unsafe.Pointer(viewport)), drawable, unsafe.Pointer(image))
	}
	return 1
}

//export nox_things_static_draw_parse
func nox_things_static_draw_parse(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char) C.bool {
	size := unsafe.Sizeof(C.nox_static_draw_data_t{})
	data := (*C.nox_static_draw_data_t)(C.nox_draw_memory_calloc(1, C.size_t(size)))
	if data == nil {
		return false
	}
	data.size = C.uint32_t(size)
	data.image = readLegacyDrawImage(file, scratch, 830856)
	thing.draw_func = (*[0]byte)(C.nox_thing_static_draw)
	thing.field_5c = unsafe.Pointer(data)
	return true
}

//export nox_things_static_random_draw_parse
func nox_things_static_random_draw_parse(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char) C.bool {
	thing.draw_func = (*[0]byte)(C.nox_thing_static_random_draw)
	data := nox_xxx_spriteLoadStaticRandomData_44C000(scratch, file)
	thing.field_5c = unsafe.Pointer(data)
	return C.bool(data != nil)
}

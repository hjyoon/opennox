package legacy

/*
#include "client__draw__animdraw.h"
#include "client__draw__canidraw.h"
#include "client__draw__parse__parse.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func readLegacyAnimationKind(file *C.nox_memfile, scratch *C.char) C.uint32_t {
	length := nox_memfile_read_u8(file)
	nox_memfile_read(unsafe.Pointer(scratch), 1, C.int(length), file)
	*(*C.char)(unsafe.Add(unsafe.Pointer(scratch), uintptr(length))) = 0
	return C.uint32_t(client.ParseAnimKind(GoString(scratch)))
}

func drawLegacyAnimationFrame(viewport *C.uint32_t, drawable *C.nox_drawable, images **nox_video_bag_image_t, count, frame int32) {
	image := unsafe.Slice(images, int(count))[frame]
	nox_xxx_drawObject_4C4770_draw((*C.nox_draw_viewport_t)(unsafe.Pointer(viewport)), drawable, unsafe.Pointer(image))
}

//export nox_thing_animate_draw
func nox_thing_animate_draw(viewport *C.uint32_t, drawable *C.nox_drawable) C.int {
	data := (*C.nox_animate_draw_data_t)(drawable.field_76)
	if data == nil || data.images == nil || data.frame_count == 0 {
		return 1
	}

	count := int32(data.frame_count)
	delay := uint32(data.frame_delay) + 1
	kind := client.AnimKind(data.animation_kind)
	var frame int32
	switch kind {
	case client.AnimOneShot:
		frame = min(int32((gameFrame()-uint32(drawable.field_79))/delay), count-1)
	case client.AnimOneShotRemove:
		frame = int32((gameFrame() - uint32(drawable.field_79)) / delay)
		if frame >= count {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(drawable)
			return 0
		}
	case client.AnimLoop:
		if drawable.flags30&0x1000000 != 0 {
			frame = int32((gameFrame()+uint32(drawable.field_32))/delay) % count
		} else if drawable.flags28&0x10000000 != 0 {
			if !noxflags.HasGame(noxflags.GameModeCTF) {
				return 1
			}
			frame = int32((gameFrame()+uint32(drawable.field_32))/delay) % count
		}
	case client.AnimLoopAndFade:
		nox_client_drawEnableAlpha_434560(1)
		frame = int32((gameFrame() - uint32(drawable.field_79)) / delay)
		if frame >= 2*count {
			nox_xxx_spriteDeleteStatic_45A4E0_drawable(drawable)
			return 0
		}
		nox_client_drawSetAlpha_434580(C.uchar(-56 - 200*frame/(2*count)))
		frame %= count
	case client.AnimRandom:
		frame = nox_common_randomIntMinMax_415FF0(0, count-1, nil, 24)
	case client.AnimSlave:
		frame = int32(drawable.field_77)
	default:
		return 1
	}
	if frame >= 0 && frame < count {
		drawLegacyAnimationFrame(viewport, drawable, data.images, count, frame)
	}
	if kind == client.AnimLoopAndFade {
		nox_client_drawEnableAlpha_434560(0)
	}
	return 1
}

//export nox_things_animate_draw_parse
func nox_things_animate_draw_parse(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char) C.bool {
	return parseAnimateDraw(thing, file, scratch, false)
}

func parseAnimateDraw(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char, high bool) C.bool {
	size := unsafe.Sizeof(C.nox_animate_draw_data_t{})
	data := (*C.nox_animate_draw_data_t)(allocateDrawMemory(1, size, high))
	if data == nil {
		return false
	}
	data.size = C.uint32_t(size)
	data.frame_count = nox_memfile_read_u8(file)
	data.frame_delay = nox_memfile_read_u8(file)
	data.animation_kind = readLegacyAnimationKind(file, scratch)
	images, ok := legacyDrawImageList(file, scratch, int(data.frame_count), 830832, high)
	if !ok {
		freeDrawMemory(unsafe.Pointer(data))
		return false
	}
	data.images = images
	thing.field_5c = unsafe.Pointer(data)
	thing.draw_func = (*[0]byte)(C.nox_thing_animate_draw)
	return true
}

//export nox_thing_cond_animate_draw
func nox_thing_cond_animate_draw(viewport *C.uint32_t, drawable *C.nox_drawable) C.int {
	data := (*C.nox_cond_animate_draw_data_t)(drawable.field_76)
	state := 1
	if drawable.flags30&0x1000000 != 0 {
		state = 0
	}
	if data == nil || data.images[state] == nil || data.frame_count[state] == 0 {
		return 0
	}
	count := int32(data.frame_count[state])
	var frame int32
	switch client.AnimKind(data.animation_kind[state]) {
	case client.AnimLoop:
		delay := uint32(data.frame_delay[state]) + 1
		frame = int32((gameFrame()+uint32(drawable.field_32))/delay) % count
	case client.AnimRandom:
		frame = nox_common_randomIntMinMax_415FF0(0, count-1, nil, 57)
	case client.AnimSlave:
		frame = int32(drawable.field_77)
	default:
		return 0
	}
	if frame < 0 || frame >= count {
		return 0
	}
	drawLegacyAnimationFrame(viewport, drawable, data.images[state], count, frame)
	return 1
}

//export nox_things_cond_animate_draw_parse
func nox_things_cond_animate_draw_parse(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char) C.bool {
	return parseConditionalAnimateDraw(thing, file, scratch, false)
}

func parseConditionalAnimateDraw(thing *C.nox_thing, file *C.nox_memfile, scratch *C.char, high bool) C.bool {
	size := unsafe.Sizeof(C.nox_cond_animate_draw_data_t{})
	data := (*C.nox_cond_animate_draw_data_t)(allocateDrawMemory(1, size, high))
	if data == nil {
		return false
	}
	data.size = C.uint32_t(size)
	states := int(nox_memfile_read_u8(file))
	if states > len(data.images) {
		freeDrawMemory(unsafe.Pointer(data))
		return false
	}
	for state := range states {
		data.frame_count[state] = nox_memfile_read_u8(file)
		data.frame_delay[state] = nox_memfile_read_u8(file)
		data.animation_kind[state] = readLegacyAnimationKind(file, scratch)
		images, ok := legacyDrawImageList(file, scratch, int(data.frame_count[state]), 830836, high)
		if !ok {
			for _, previous := range data.images[:state] {
				freeDrawMemory(unsafe.Pointer(previous))
			}
			freeDrawMemory(unsafe.Pointer(data))
			return false
		}
		data.images[state] = images
	}
	thing.field_5c = unsafe.Pointer(data)
	thing.draw_func = (*[0]byte)(C.nox_thing_cond_animate_draw)
	thing.field_60 = 0
	return true
}

//export nox_things_animate_state_draw_parse
func nox_things_animate_state_draw_parse(thing *C.nox_thing, file *C.nox_memfile, _ *C.char) C.bool {
	size := unsafe.Sizeof(client.AnimationStateDrawData{})
	data := (*client.AnimationStateDrawData)(allocateDrawMemory(1, size, false))
	if data == nil {
		return false
	}
	resolve := func(id int, kind byte, name string) noxrender.ImageHandle {
		return GetClient().R2().GetBag().ImageRef(id, kind, name).C()
	}
	if err := client.ParseAnimStateDrawDataWithAllocator(data, asMemfile(file), resolve, legacyAllocateAnimationFrames); err != nil {
		freeDrawMemory(unsafe.Pointer(data))
		return false
	}
	thing.field_54 = 2
	thing.draw_func = (*[0]byte)(C.nox_thing_animate_state_draw)
	thing.field_5c = unsafe.Pointer(data)
	return true
}

func legacyAllocateAnimationFrames(count int) ([]noxrender.ImageHandle, func()) {
	pointer := allocateDrawMemory(count, unsafe.Sizeof(noxrender.ImageHandle(nil)), false)
	if pointer == nil {
		return nil, nil
	}
	return unsafe.Slice((*noxrender.ImageHandle)(pointer), count), func() { freeDrawMemory(pointer) }
}

//export nox_xxx_loadVectorAnimated_44B8B0
func nox_xxx_loadVectorAnimated_44B8B0(pointer unsafe.Pointer, file *C.nox_memfile) C.int {
	vector := (*client.AnimationVector)(pointer)
	memfile := asMemfile(file)
	vector.Cnt40 = uint16(memfile.ReadU8())
	vector.Val42 = uint16(memfile.ReadU8())
	name, err := memfile.ReadString8()
	if err != nil {
		return 0
	}
	vector.Kind = client.ParseAnimKind(name)
	return 1
}

//export sub_44BE90
func sub_44BE90(pointer unsafe.Pointer, file *C.nox_memfile) C.int {
	vector := (*client.AnimationVector)(pointer)
	var scratch [256]C.char
	images, ok := legacyDrawImageList(file, &scratch[0], int(vector.Cnt40), 830848, false)
	if !ok {
		return 0
	}
	vector.Frames[0] = (*noxrender.ImageHandle)(unsafe.Pointer(images))
	return 1
}

//export nox_xxx_loadVectorAnimated_44BC50
func nox_xxx_loadVectorAnimated_44BC50(pointer unsafe.Pointer, file *C.nox_memfile) C.int {
	vector := (*client.AnimationVector)(pointer)
	var frames [9]*noxrender.ImageHandle
	var scratch [256]C.char
	for direction := range frames {
		if direction == 4 {
			continue
		}
		images, ok := legacyDrawImageList(file, &scratch[0], int(vector.Cnt40), 830844, false)
		if !ok {
			for _, frame := range frames[:direction] {
				freeDrawMemory(unsafe.Pointer(frame))
			}
			return 0
		}
		frames[direction] = (*noxrender.ImageHandle)(unsafe.Pointer(images))
	}
	frames[4] = vector.Frames[4]
	vector.Frames = frames
	return 1
}

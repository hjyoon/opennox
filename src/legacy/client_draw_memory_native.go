package legacy

/*
#include "defs.h"
#include "client__draw__parse__parse.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func freeLegacyVectorFrames(frames *[9]*noxrender.ImageHandle) {
	for direction, images := range frames {
		if direction == 4 || images == nil {
			continue
		}
		nox_draw_memory_release_go(unsafe.Pointer(images))
		frames[direction] = nil
	}
}

func freeLegacyEquipmentFrames(animation *client.PlayerEquipAnimation) {
	if animation == nil {
		return
	}
	freeLegacyVectorFrames(&animation.Frames)
	nox_draw_memory_release_go(unsafe.Pointer(animation))
}

func freeLegacyPlayerDrawFrames(data *client.PlayerDrawData) {
	for i := range data.Anim {
		animation := &data.Anim[i]
		freeLegacyVectorFrames(&animation.Base.Frames)
		freeLegacyEquipmentFrames(animation.Naked)
		for _, equipment := range animation.Armor {
			freeLegacyEquipmentFrames(equipment)
		}
		for _, equipment := range animation.Weapon {
			freeLegacyEquipmentFrames(equipment)
		}
	}
}

//export nox_xxx_draw_44C650_free_kind
func nox_xxx_draw_44C650_free_kind(pointer unsafe.Pointer, kind C.int) {
	if pointer == nil {
		return
	}
	switch kind {
	case 2:
		data := (*C.nox_static_random_draw_data_t)(pointer)
		nox_draw_memory_release_go(unsafe.Pointer(data.images))
	case 3:
		data := (*C.nox_animate_draw_data_t)(pointer)
		nox_draw_memory_release_go(unsafe.Pointer(data.images))
	case 4:
		data := (*C.nox_cond_animate_draw_data_t)(pointer)
		for _, images := range data.images {
			nox_draw_memory_release_go(unsafe.Pointer(images))
		}
	case 5:
		data := (*client.AnimationVector)(pointer)
		freeLegacyVectorFrames(&data.Frames)
	case 6:
		freeLegacyPlayerDrawFrames((*client.PlayerDrawData)(pointer))
	case 7:
		data := (*client.MonsterDrawData)(pointer)
		for i := range data.Anim {
			freeLegacyVectorFrames(&data.Anim[i].Frames)
		}
	case 8:
		data := (*client.AnimationStateDrawData)(pointer)
		for i := range data.Anim {
			freeLegacyVectorFrames(&data.Anim[i].Frames)
		}
	}
	nox_draw_memory_release_go(pointer)
}

//export nox_free_tile_defs
func nox_free_tile_defs() C.int {
	tiles := nox_floor_tile_storage()
	count := min(int(tiles.count), len(tiles.definitions))
	for i := range count {
		tile := &tiles.definitions[i]
		if tile.data_32 != nil {
			nox_draw_memory_release_go(unsafe.Pointer(tile.data_32))
			tile.data_32 = nil
		}
	}
	return 1
}

//export sub_485F30
func sub_485F30() C.int {
	count := min(int(int32(*thingEdgeCountCell411850())), len(C.nox_edge_images_native))
	for i := range count {
		images := thingEdgeImageCell411850(uint32(i))
		if *images != nil {
			nox_draw_memory_release_go(*images)
			*images = nil
		}
		thingEdgeStore32411850(uint32(i), 28676, 0)
	}
	return 1
}

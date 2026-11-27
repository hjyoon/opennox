package legacy

/*
#include "native_callback_calls.h"
*/
import "C"

import "unsafe"

func callFloorTileDraw(draw unsafe.Pointer, pos *C.int2, image *C.nox_video_bag_image_t, tile C.uint16_t) {
	C.nox_legacy_call_floor_tile(C.nox_floor_tile_draw_fn(draw), pos, image, tile)
}

func callFloorEdgeDraw(draw unsafe.Pointer, pos *C.int2, edge *C.nox_tile_list_node_t) {
	C.nox_legacy_call_floor_edge(C.nox_floor_edge_draw_fn(draw), pos, edge)
}

func callObjectTypeParser(parse unsafe.Pointer, text *C.char, target unsafe.Pointer) int {
	return int(C.nox_legacy_call_object_parser((*[0]byte)(parse), text, target))
}

func callObjectTypeConstructor(init unsafe.Pointer, object *C.nox_object_t) {
	C.nox_legacy_call_object_constructor((*[0]byte)(init), object)
}

func callDrawableVisitor(visit unsafe.Pointer, drawable *C.nox_drawable, data unsafe.Pointer) {
	C.nox_legacy_call_drawable_visitor((*[0]byte)(visit), drawable, data)
}

func callUnitRadiusCallback(visit C.nox_unit_radius_callback_t, candidate *C.nox_object_t, data unsafe.Pointer) {
	C.nox_legacy_call_unit_radius(visit, candidate, data)
}

func callPrefabConnector(connect C.nox_mapgen_prefab_connect_cb_5266f0, prefab *C.uint8_t, direction C.int, candidate *C.uint8_t) C.int {
	return C.nox_legacy_call_prefab_connector(connect, prefab, direction, candidate)
}

func callSpellC(cast unsafe.Pointer, spell C.int, first, second, third *C.nox_object_t, data unsafe.Pointer, level C.int) int {
	return int(C.nox_legacy_call_spell((*[0]byte)(cast), spell, first, second, third, data, level))
}

func callWorldDefend(callback, effect unsafe.Pointer, item *C.nox_object_t, value *C.float) {
	C.nox_legacy_call_world_defend(callback, effect, item, 0, item, 0, value)
}

func callMapPushBefore(callback C.uintptr_t, candidate *C.nox_object_t, distance C.uint32_t, arg C.uintptr_t) {
	C.nox_legacy_call_map_push_before(callback, candidate, distance, arg)
}

func callItemAmountCallback(callback C.nox_gui_item_amount_callback_t, pos *C.int2, id, thing, amount, extra C.uint32_t) {
	C.nox_legacy_call_item_amount(callback, pos, id, thing, amount, extra)
}

#ifndef OPENNOX_NATIVE_CALLBACK_CALLS_H
#define OPENNOX_NATIVE_CALLBACK_CALLS_H

#include "defs.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "native_call_abi.h"

// Keep the legacy C parameter types at the cgo boundary. The shared frames
// accept the corresponding integer and pointer slots in native ABI order.
typedef void (*nox_floor_tile_draw_fn)(int2*, nox_video_bag_image_t*, uint16_t);
typedef void (*nox_floor_edge_draw_fn)(int2*, nox_tile_list_node_t*);

NOX_NATIVE_CALL(void, nox_legacy_call_floor_tile, nox_ccall_void_ptr3,
    (nox_floor_tile_draw_fn draw, int2* pos, nox_video_bag_image_t* image, uint16_t tile),
    draw(pos, image, tile);)
NOX_NATIVE_CALL(void, nox_legacy_call_floor_edge, nox_ccall_void_ptr2,
    (nox_floor_edge_draw_fn draw, int2* pos, nox_tile_list_node_t* edge),
    draw(pos, edge);)
NOX_NATIVE_CALL(int, nox_legacy_call_object_parser, nox_ccall_int_ptr2,
    (int (*parse)(char*, void*), char* text, void* target),
    return parse(text, target);)
NOX_NATIVE_CALL(void, nox_legacy_call_object_constructor, nox_ccall_void_ptr,
    (void (*init)(nox_object_t*), nox_object_t* object),
    init(object);)
NOX_NATIVE_CALL(void, nox_legacy_call_drawable_visitor, nox_ccall_void_ptr2,
    (void (*visit)(nox_drawable*, void*), nox_drawable* drawable, void* data),
    visit(drawable, data);)
NOX_NATIVE_CALL(void, nox_legacy_call_unit_radius, nox_ccall_void_ptr2,
    (nox_unit_radius_callback_t visit, nox_object_t* candidate, void* data),
    visit(candidate, data);)
NOX_NATIVE_CALL(int, nox_legacy_call_prefab_connector, nox_ccall_int_ptr3,
    (nox_mapgen_prefab_connect_cb_5266f0 connect, uint8_t* prefab, int direction, uint8_t* candidate),
    return connect(prefab, direction, candidate);)
NOX_NATIVE_CALL(int, nox_legacy_call_spell, nox_ccall_int_ptr6,
    (int (*cast)(int, void*, nox_object_t*, nox_object_t*, void*, int),
     int spell, nox_object_t* first, nox_object_t* second, nox_object_t* third, void* data, int level),
    return cast(spell, first, second, third, data, level);)
NOX_NATIVE_CALL(void, nox_legacy_call_world_defend, nox_ccall_void_ptr6,
    (void* callback, void* effect, nox_object_t* source, uintptr_t source_arg,
     nox_object_t* target, uintptr_t target_arg, float* value),
    ((void (*)(void*, nox_object_t*, uintptr_t, nox_object_t*, uintptr_t, float*))callback)(
        effect, source, source_arg, target, target_arg, value);)
NOX_NATIVE_CALL(int32_t, nox_legacy_call_modifier_predicate, nox_ccall_int_ptr6,
    (void* callback, void* effect, uintptr_t item, uintptr_t context,
     uintptr_t target, uintptr_t other, int32_t* output),
    return ((int32_t (*)(void*, nox_object_t*, uintptr_t, nox_object_t*, uintptr_t, int32_t*))callback)(
        effect, (nox_object_t*)item, context, (nox_object_t*)target, other, output);)
NOX_NATIVE_CALL(void, nox_legacy_call_map_push_before, nox_ccall_void_ptr3,
    (uintptr_t callback, nox_object_t* candidate, uint32_t distance, uintptr_t arg),
    ((void (*)(nox_object_t*, uint32_t, uintptr_t))callback)(candidate, distance, arg);)
NOX_NATIVE_CALL(void, nox_legacy_call_item_amount, nox_ccall_void_ptr5,
    (nox_gui_item_amount_callback_t callback, int2* pos, uint32_t id,
     uint32_t thing, uint32_t amount, uint32_t extra),
    callback(pos, id, thing, amount, extra);)

#endif

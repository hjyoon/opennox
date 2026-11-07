#ifndef NOX_PORT_RAW_DEFAULT_DAMAGE_NATIVE_BRIDGE_4E0B30_H
#define NOX_PORT_RAW_DEFAULT_DAMAGE_NATIVE_BRIDGE_4E0B30_H

#include <stdint.h>

typedef struct nox_object_t nox_object_t;

// Object arguments stay pointer-valued void* at the cgo boundary so a narrow
// or unaligned foreign fixture is not cast to a full Go C-struct view first.
// The actual C callback still receives the declared nox_object_t* arguments.
void rdd_native_damage_callback_4e0b30(
    int (*callback)(nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t),
    void* target, void* source, void* weapon, int32_t damage, int32_t typ);
void rdd_native_sound_callback_4e0b30(
    int (*callback)(nox_object_t*, nox_object_t*), void* target, void* source);

uint32_t rdd_native_game_flags_word_4e0b30(uint32_t flags);
uint32_t rdd_native_gameplay_flags_word_4e0b30(int32_t flags);

#endif

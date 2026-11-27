//go:build !386 && !amd64 && !arm && !arm64

#include "raw_default_damage_native_bridge_4e0b30.h"

typedef int (*rdd_native_damage_function_4e0b30)(
    nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef int (*rdd_native_sound_function_4e0b30)(nox_object_t*, nox_object_t*);

#if defined(__clang__)
#define RDD_NATIVE_CALLBACK_FRAME_4E0B30 __attribute__((noinline, disable_tail_calls))
#elif defined(__GNUC__)
#define RDD_NATIVE_CALLBACK_FRAME_4E0B30 __attribute__((noinline))
#else
#error "native damage callback frame requires the project's GNU-compatible C compiler"
#endif

// Keep the indirect call in this named frame on the fallback targets.
RDD_NATIVE_CALLBACK_FRAME_4E0B30
void rdd_native_damage_callback_4e0b30(
    rdd_native_damage_function_4e0b30 callback,
    void* target, void* source, void* weapon, int32_t damage, int32_t typ) {
    // No nil/unknown admission guard: an original attempted call must execute.
    (void)callback((nox_object_t*)target, (nox_object_t*)source,
                   (nox_object_t*)weapon, damage, typ);
    __asm__ __volatile__("" ::: "memory");
}

RDD_NATIVE_CALLBACK_FRAME_4E0B30
void rdd_native_sound_callback_4e0b30(
    rdd_native_sound_function_4e0b30 callback, void* target, void* source) {
    // Product DefaultDamageSound/PlayerDamageSound return int; the parent
    // ignores that result. Known fixture void endpoints are handled by the
    // declared observer before this actual product bridge is reached.
    (void)callback((nox_object_t*)target, (nox_object_t*)source);
    __asm__ __volatile__("" ::: "memory");
}

#undef RDD_NATIVE_CALLBACK_FRAME_4E0B30

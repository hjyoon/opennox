#include "raw_default_damage_native_bridge_4e0b30.h"

#include <limits.h>
#include "GAME1.h"
#include "GAME4_3.h"

typedef int (*rdd_native_damage_function_4e0b30)(
    nox_object_t*, nox_object_t*, nox_object_t*, int32_t, int32_t);
typedef int (*rdd_native_sound_function_4e0b30)(nox_object_t*, nox_object_t*);

_Static_assert(CHAR_BIT == 8 && sizeof(int) == 4 && sizeof(unsigned int) == 4,
               "native damage bridge requires the existing 32-bit C integer ABI");
_Static_assert(_Generic(((nox_object_t*)0)->func_damage,
                        rdd_native_damage_function_4e0b30: 1, default: 0),
               "native damage callback must retain three pointers and two int32 arguments");
_Static_assert(_Generic(&nox_xxx_soundDefaultDamageSound_532E20,
                        rdd_native_sound_function_4e0b30: 1, default: 0),
               "registered product sound callback returns int");
_Static_assert(_Generic(&nox_common_gameFlags_check_40A5C0,
                        bool (*)(unsigned int): 1, default: 0),
               "game flags service retains C bool return");
_Static_assert(_Generic(&nox_xxx_CheckGameplayFlags_417DA0,
                        bool (*)(int): 1, default: 0),
               "gameplay flags service retains C bool return");

#if defined(__clang__)
#define RDD_NATIVE_CALLBACK_FRAME_4E0B30 __attribute__((noinline, disable_tail_calls))
#elif defined(__GNUC__)
#define RDD_NATIVE_CALLBACK_FRAME_4E0B30 __attribute__((noinline))
#else
#error "native damage callback frame requires the project's GNU-compatible C compiler"
#endif

// One definition per production translation unit. The noinline attribute and
// post-call compiler barrier keep the indirect call in this named frame. The
// retained ARM64 binary must still independently prove its actual LR-4 BLR.
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

uint32_t rdd_native_game_flags_word_4e0b30(uint32_t flags) {
    return nox_common_gameFlags_check_40A5C0((unsigned int)flags)
               ? UINT32_C(1) : UINT32_C(0);
}

uint32_t rdd_native_gameplay_flags_word_4e0b30(int32_t flags) {
    return nox_xxx_CheckGameplayFlags_417DA0((int)flags)
               ? UINT32_C(1) : UINT32_C(0);
}

#undef RDD_NATIVE_CALLBACK_FRAME_4E0B30

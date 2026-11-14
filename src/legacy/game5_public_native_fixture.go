package legacy

/*
#include <stddef.h>
#include <limits.h>
#include <string.h>
#include "GAME5_2.h"
static size_t game5_public_player_flags_offset(void) { return offsetof(nox_playerInfo, field_3680); }
static int game5_public_can_talk(void *player) { return nox_xxx_playerCanTalkMB_57A160(player); }
static uintptr_t game5_public_reflect(void *normal, void *velocity) {
    return (uintptr_t)nox_xxx_collideReflect_57B810(normal, velocity);
}
static int game5_public_x87_available(void) {
#if (defined(__i386__) || defined(__x86_64__)) && defined(__GNUC__)
    return 1;
#else
    return 0;
#endif
}
// Independent ABI adaptation of the original x87/MOV instruction sequence.
// The sealed Unicorn fixture stays unchanged: Unicorn 2.1.4 does not quiet
// signaling NaNs on FLD, as the existing velocity-load probe also establishes.
static int game5_public_x87_reflect(void *normal, void *velocity, void *output, int delta) {
#if (defined(__i386__) || defined(__x86_64__)) && defined(__GNUC__)
    unsigned char owner[32], saved[28];
    uint32_t normal_bits[2];
    static const float zero;
    uint16_t cw = 0xe3f;
    memcpy(owner, (unsigned char *)velocity - 8, sizeof(owner));
    memcpy(normal_bits, normal, sizeof(normal_bits));
    const float *n = (const float *)normal_bits;
    if (delta != INT_MIN) {
        if (delta != -4 && delta != 0 && delta != 4) return -1;
        n = (const float *)(owner + 8 + delta);
    }
    float *v = (float *)(owner + 8);
    __asm__ volatile(
        "fnstenv %[saved]\n\tfnclex\n\tfldcw %[cw]\n\t"
        "flds 4(%[normal])\n\tfmuls (%[normal])\n\tfcomps %[zero]\n\t"
        "fnstsw %%ax\n\ttestb $0x41, %%ah\n\tflds (%[velocity])\n\tjne 1f\n\t"
        "flds 4(%[velocity])\n\tfchs\n\tfstps (%[velocity])\n\t"
        "fchs\n\tfstps 4(%[velocity])\n\tjmp 2f\n\t"
        "1: movl 4(%[velocity]), %%ecx\n\tfstps 4(%[velocity])\n\tmovl %%ecx, (%[velocity])\n\t"
        "2: fldenv %[saved]\n\t"
        : [saved] "=m" (saved)
        : [normal] "r" (n), [velocity] "r" (v), [zero] "m" (zero), [cw] "m" (cw)
        : "ax", "cx", "cc", "memory", "st", "st(1)");
    memcpy(output, owner, sizeof(owner));
    return 1;
#else
    (void)normal; (void)velocity; (void)output; (void)delta;
    return 0;
#endif
}
*/
import "C"

import "unsafe"

func game5PublicPlayerLayoutC() (uintptr, uintptr) {
	return C.sizeof_nox_playerInfo, uintptr(C.game5_public_player_flags_offset())
}
func game5PublicCanTalkC(player unsafe.Pointer) int32 {
	return int32(C.game5_public_can_talk(player))
}
func game5PublicReflectC(normal, velocity unsafe.Pointer) uintptr {
	return uintptr(C.game5_public_reflect(normal, velocity))
}
func game5PublicX87AvailableC() bool { return C.game5_public_x87_available() != 0 }
func game5PublicX87ReflectC(normal, velocity unsafe.Pointer, output *[32]byte, delta *int) bool {
	d := int32(-2147483648)
	if delta != nil {
		d = int32(*delta)
	}
	return C.game5_public_x87_reflect(normal, velocity, unsafe.Pointer(&output[0]), C.int(d)) == 1
}

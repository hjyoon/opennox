#include "GAME3.h"
#include <stdint.h>

/* The preserved interfaces return mixed opaque scalar/pointer domains.
 * Keep those ABI bits in C; never manufacture a Go pointer from a frame or 1. */
extern uintptr_t game3_summon_packet_native(short, unsigned short*, unsigned short, unsigned char, short);
extern uintptr_t game3_fx_shield_native(unsigned int, int);
extern uintptr_t game3_equipment_materials_native(void*, uintptr_t);

uint32_t* nox_xxx_netHandleSummonPacket_4B7C40(short id, unsigned short* xy, unsigned short thing,
                                           unsigned char direction, short lifetime) {
    return (uint32_t*)game3_summon_packet_native(id, xy, thing, direction, lifetime);
}
uint32_t* nox_xxx_fxShield_4B8090(unsigned int code, int direction) {
    return (uint32_t*)game3_fx_shield_native(code, direction);
}
uint32_t* sub_4B8E10(uint32_t* entries, char* token) {
    return (uint32_t*)game3_equipment_materials_native(entries, (uintptr_t)token);
}

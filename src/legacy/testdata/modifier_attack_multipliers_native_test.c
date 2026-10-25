#include <assert.h>
#include <math.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#undef _Static_assert
_Static_assert(sizeof(void*) == 8, "native pointer width");
_Static_assert(sizeof(float) == 4, "original float width");
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME3_2.h"
static struct { uint32_t before[8]; float coefficient; uint32_t after[8]; } effect;
static struct { uint32_t before; float value; uint32_t after; } output;
static nox_object_t weapon, owner, target, projectile;
static int getter_calls;
float nox_modifier_effect_getAttackFloat(void* ptr) {
    assert(ptr == &effect); getter_calls++; return effect.coefficient;
}
// PRODUCTION_DOUBLE2FLOAT_419AF0
// PRODUCTION_MUL_CHOP_4E0397
// PRODUCTION_DAMAGE_MULTIPLIER_4E04C0
// PRODUCTION_PROJECTILE_SPEED_4E09B0
int main(int argc, char** argv) {
    assert(argc == 4);
    int which = atoi(argv[1]); assert(which == 0 || which == 1);
    memset(&effect, 0x5a, sizeof(effect)); memset(&output, 0xa5, sizeof(output));
    memset(&weapon, 0x96, sizeof(weapon)); memset(&owner, 0x69, sizeof(owner));
    memset(&target, 0x3c, sizeof(target)); memset(&projectile, 0xc3, sizeof(projectile));
    uint32_t value_bits = (uint32_t)strtoul(argv[2], NULL, 0);
    uint32_t coefficient_bits = (uint32_t)strtoul(argv[3], NULL, 0);
    memcpy(&output.value, &value_bits, 4); memcpy(&effect.coefficient, &coefficient_bits, 4);
    memcpy(&projectile.speed_cur, &value_bits, 4);
    unsigned char before_effect[sizeof(effect)], before_weapon[sizeof(weapon)], before_owner[sizeof(owner)], before_target[sizeof(target)], before_projectile[sizeof(projectile)];
    memcpy(before_effect, &effect, sizeof(effect)); memcpy(before_weapon, &weapon, sizeof(weapon));
    memcpy(before_owner, &owner, sizeof(owner)); memcpy(before_target, &target, sizeof(target));
    memcpy(before_projectile, &projectile, sizeof(projectile));
    unsigned char before_output[sizeof(output)]; memcpy(before_output,&output,sizeof(output));
    uintptr_t ptrs[] = {(uintptr_t)&effect,(uintptr_t)&weapon,(uintptr_t)&owner,(uintptr_t)&target,(uintptr_t)&output.value,(uintptr_t)&projectile};
    for (int i=0;i<6;i++) assert(ptrs[i] > UINT32_MAX);
    if (which == 0) {
        float* result = nox_xxx_effectDamageMultiplier_4E04C0(&effect,&weapon,&owner,&target,&output.value);
        assert(result == &output.value); memcpy(&value_bits, &output.value, 4);
        memcpy(&output.value,before_output+sizeof(uint32_t),4);
    } else {
        nox_object_t* result = nox_xxx_effectProjectileSpeed_4E09B0(&effect,&weapon,&owner,&target,&projectile);
        assert(result == &projectile); memcpy(&value_bits, &projectile.speed_cur, 4);
        memcpy(&projectile.speed_cur, before_projectile+offsetof(nox_object_t,speed_cur), 4);
    }
    assert(getter_calls == 1 && !memcmp(before_output,&output,sizeof(output)));
    assert(!memcmp(before_effect,&effect,sizeof(effect)) && !memcmp(before_weapon,&weapon,sizeof(weapon)) && !memcmp(before_owner,&owner,sizeof(owner)) && !memcmp(before_target,&target,sizeof(target)) && !memcmp(before_projectile,&projectile,sizeof(projectile)));
    assert(output.before == UINT32_C(0xa5a5a5a5) && output.after == UINT32_C(0xa5a5a5a5));
    printf("{\"value_bits\":%u,\"return_pointer\":\"%s\",\"getter_calls\":1,\"high_native_pointers\":true,\"nonoutput_canaries_unchanged\":true}\n", value_bits,which ? "projectile" : "value");
    return 0;
}

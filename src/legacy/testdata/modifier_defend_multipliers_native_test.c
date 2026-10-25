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
static nox_object_t item, target;
static int getter_calls;
float nox_modifier_effect_getDefendFloat(void* ptr) {
    assert(ptr == &effect);
    getter_calls++;
    return effect.coefficient;
}
// PRODUCTION_DOUBLE2FLOAT_419AF0
// PRODUCTION_MULTIPLIERS_4E0370_4E0380
int main(int argc, char** argv) {
    assert(argc == 4);
    int which = atoi(argv[1]); assert(which == 0 || which == 1);
    memset(&effect, 0x5a, sizeof(effect)); memset(&output, 0xa5, sizeof(output));
    uint32_t value_bits = (uint32_t)strtoul(argv[2], NULL, 0);
    uint32_t coefficient_bits = (uint32_t)strtoul(argv[3], NULL, 0);
    memcpy(&output.value, &value_bits, 4); memcpy(&effect.coefficient, &coefficient_bits, 4);
    unsigned char before_effect[sizeof(effect)], before_item[sizeof(item)], before_target[sizeof(target)];
    memcpy(before_effect, &effect, sizeof(effect)); memcpy(before_item, &item, sizeof(item)); memcpy(before_target, &target, sizeof(target));
    uintptr_t ptrs[] = {(uintptr_t)&effect,(uintptr_t)&item,(uintptr_t)&target,(uintptr_t)&output,(uintptr_t)&output.value};
    for (int i=0;i<5;i++) assert(ptrs[i] > UINT32_MAX);
    float* result = which ? sub_4E0380(&effect, &item, (uintptr_t)&item, &target, (uintptr_t)&target, &output.value)
                          : sub_4E0370(&effect, &item, (uintptr_t)&item, &target, (uintptr_t)&target, &output.value);
    assert(result == &output.value && getter_calls == 1);
    assert(!memcmp(before_effect, &effect, sizeof(effect)) && !memcmp(before_item, &item, sizeof(item)) && !memcmp(before_target, &target, sizeof(target)));
    assert(output.before == UINT32_C(0xa5a5a5a5) && output.after == UINT32_C(0xa5a5a5a5));
    memcpy(&value_bits, &output.value, 4);
    printf("{\"value_bits\":%u,\"return_pointer\":\"value\",\"getter_calls\":1,"
           "\"high_native_pointers\":true,\"nonoutput_canaries_unchanged\":true}\n",value_bits);
    return 0;
}
